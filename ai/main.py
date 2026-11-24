import json
import logging
import os
import random
import uuid
from pathlib import Path
from typing import Any, Dict, List, Optional

import pandas as pd
from catboost import CatBoostClassifier
from fastapi import Body, FastAPI, HTTPException
from pydantic import BaseModel, Field

from plan_service import create_router

logger = logging.getLogger("ai_gym_coach_ml")
logging.basicConfig(level=logging.INFO)

BASE_DIR = Path(__file__).resolve().parent

programs: List[Dict[str, Any]] = []
_dataset_path = BASE_DIR / "dataset/dataset.jsonlines"
if _dataset_path.exists():
    with _dataset_path.open("r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if line:
                programs.append(json.loads(line)["output"])

app = FastAPI(title="AI-Gym-Coach ML Service", version="v2")

reverse_map = {
    0: "Фулбади",
    1: "ВерхНиз",
    2: "Тяни-толкай",
    3: "Пуш-пул-ноги",
    4: "Сплит",
}

_model_candidates = [
    BASE_DIR / "models/model/workout_model.cbm",
    BASE_DIR / "workout_model.cbm",
]
CATBOOST_MODEL_PATH: Optional[Path] = next((p for p in _model_candidates if p.exists()), None)
_catboost_model: Optional[CatBoostClassifier] = None

if CATBOOST_MODEL_PATH is not None:
    try:
        _m = CatBoostClassifier()
        _m.load_model(str(CATBOOST_MODEL_PATH))
        _catboost_model = _m
        logger.info("Загружена модель CatBoost из %s", CATBOOST_MODEL_PATH)
    except Exception as exc:
        logger.warning(
            "Не удалось загрузить %s (%s); используем резервную логику.",
            CATBOOST_MODEL_PATH,
            exc,
        )

model = _catboost_model


def get_random_program(programs: List[Dict[str, Any]], split_name: str, days: List[Any]) -> Optional[Dict[str, Any]]:
    """Возвращает случайную программу тренировок из датасета по предсказанному сплиту и числу дней."""
    filtered = [
        p
        for p in programs
        if p.get("тип_сплита") == split_name and len(p.get("еженедельный_план", [])) == len(days)
    ]
    if not filtered:
        return None
    return random.choice(filtered)


def predict(data: Dict[str, Any]):
    """Legacy-предсказание сплита и списка дней по профилю пользователя."""
    mapped = {
        "возраст": data.get("age", data.get("возраст", 30)),
        "рост": data.get("height_cm", data.get("рост", 175)),
        "вес": data.get("weight_kg", data.get("вес", 75)),
    }

    gender = str(data.get("gender", data.get("пол", "м"))).lower()
    mapped["пол"] = "ж" if any(x in gender for x in ["жен", "female", "ж"]) else "м"

    activity = data.get("activity_level", data.get("тип_активности", "Средняя активность"))
    if isinstance(activity, str) and "(" in activity:
        activity = activity.split("(")[0].strip()
    mapped["тип_активности"] = activity

    injuries = data.get("injuries_notes", data.get("травмы_или_болезни", False))
    if isinstance(injuries, bool):
        mapped["травмы_или_болезни"] = "да" if injuries else "нет"
    else:
        mapped["травмы_или_болезни"] = "да" if str(injuries).lower() in ["true", "да", "1", "yes"] else "нет"

    goal = data.get("goal", data.get("цель", "Набрать мышцы"))
    if goal == "Сбросить вес":
        goal = "Скинуть вес"
    mapped["цель"] = goal

    fitness = str(data.get("fitness_level", data.get("уровень_подготовки", "любитель"))).lower()
    if any(x in fitness for x in ["новичок", "beginner"]):
        mapped["уровень_подготовки"] = "новичок"
    elif any(x in fitness for x in ["продвинутый", "advanced"]):
        mapped["уровень_подготовки"] = "продвинутый"
    else:
        mapped["уровень_подготовки"] = "любитель"

    days = data.get("training_days_map", data.get("дни_тренировок", ["Пн", "Ср", "Пт"]))

    df = pd.DataFrame([mapped])
    df["пол_num"] = df["пол"].map({"м": 0, "ж": 1})
    df["травмы_num"] = df["травмы_или_болезни"].map({"нет": 0, "да": 1})
    df["уровень_num"] = df["уровень_подготовки"].map({"новичок": 0, "любитель": 1, "продвинутый": 2})

    features = df[
        ["возраст", "рост", "вес", "пол_num", "тип_активности", "травмы_num", "цель", "уровень_num"]
    ]

    if _catboost_model is not None:
        pred = _catboost_model.predict(features)[0][0]
        pred_idx = int(pred)
        return reverse_map.get(pred_idx, "Фулбади"), days

    fallback_split = "Фулбади" if len(days) <= 3 else ("ВерхНиз" if len(days) == 4 else "Пуш-пул-ноги")
    return fallback_split, days


class Profile(BaseModel):
    age: int = Field(..., ge=10, le=100)
    height_cm: float = Field(..., ge=100, le=250)
    weight_kg: float = Field(..., ge=30, le=300)
    level: str
    goal: str
    training_days: List[int]
    equipment_ids: List[str] = []
    excluded_exercise_ids: List[int] = []


class History(BaseModel):
    completed_28d: Optional[int] = 0
    adherence_28d: Optional[float] = 0.0
    exercise_metrics: Optional[List[Any]] = []


class GeneratePlanRequest(BaseModel):
    request_id: Optional[str] = None
    feature_schema_version: str = "v2.0"
    profile: Profile
    history: Optional[History] = None


class ExerciseItem(BaseModel):
    exercise_id: int
    sets: int
    reps_min: int
    reps_max: int
    weight_kg: Optional[float] = None
    rest_sec: int


class WorkoutDay(BaseModel):
    weekday: int
    title: str
    exercises: List[ExerciseItem]


class GeneratePlanResponse(BaseModel):
    request_id: str
    model_version: str
    feature_schema_version: str
    split: str
    days: List[WorkoutDay]
    warnings: List[str] = []
    fallback: bool = False


# Каталог упражнений: (exercise_id, sets, reps_min, reps_max, rest_sec)
EXERCISE_CATALOG: List[tuple[int, int, int, int, int]] = [
    (46, 3, 8, 12, 120),
    (101, 3, 10, 15, 90),
    (102, 3, 10, 15, 90),
    (103, 3, 12, 15, 75),
    (104, 3, 12, 15, 60),
]

# Минимальный безопасный каталог для fallback: (exercise_id, sets, reps_min, reps_max)
SAFE_FALLBACK_CATALOG: List[tuple[int, int, int, int]] = [
    (46, 2, 10, 12),
    (101, 2, 10, 15),
    (103, 2, 12, 15),
]


def predict_split_label(payload: GeneratePlanRequest) -> str:
    """Определяет сплит через CatBoost при наличии модели или по числу дней тренировок."""
    days_count = len(payload.profile.training_days)
    if _catboost_model is not None:
        try:
            level_map = {"новичок": 0, "любитель": 1, "продвинутый": 2}
            lvl_num = level_map.get(payload.profile.level.strip().lower(), 0)
            df_x = pd.DataFrame(
                [
                    {
                        "возраст": payload.profile.age,
                        "рост": payload.profile.height_cm,
                        "вес": payload.profile.weight_kg,
                        "пол_num": 0,
                        "тип_активности": "Умеренная активность",
                        "травмы_num": 1 if payload.profile.excluded_exercise_ids else 0,
                        "цель": payload.profile.goal,
                        "уровень_num": lvl_num,
                    }
                ]
            )
            pred_id = int(_catboost_model.predict(df_x).flatten()[0])
            id_to_split = {
                0: "full_body",
                1: "upper_lower",
                2: "push_pull",
                3: "push_pull_legs",
                4: "split",
            }
            return id_to_split.get(pred_id, "full_body")
        except Exception as exc:
            logger.warning("Ошибка предсказания CatBoost (%s); используем правило по дням.", exc)

    if days_count <= 3:
        return "full_body"
    if days_count == 4:
        return "upper_lower"
    return "push_pull_legs"


def build_fallback_plan(
    payload: GeneratePlanRequest,
    weekdays: List[int],
    reasons: List[str],
) -> GeneratePlanResponse:
    """Детерминированный безопасный план при сбое модели или нарушении ограничений."""
    excluded = set(payload.profile.excluded_exercise_ids or [])
    allowed_exercises = [ex for ex in SAFE_FALLBACK_CATALOG if ex[0] not in excluded]
    if not allowed_exercises:
        allowed_exercises = [(46, 2, 10, 12)]

    rest_sec = 90
    days: List[WorkoutDay] = []
    for idx, day_num in enumerate(weekdays):
        ex_id, sets, reps_min, reps_max = allowed_exercises[idx % len(allowed_exercises)]
        days.append(
            WorkoutDay(
                weekday=day_num,
                title=f"День {chr(65 + (idx % 26))} (безопасный план)",
                exercises=[
                    ExerciseItem(
                        exercise_id=ex_id,
                        sets=sets,
                        reps_min=reps_min,
                        reps_max=reps_max,
                        weight_kg=None,
                        rest_sec=rest_sec,
                    )
                ],
            )
        )

    warnings = ["FALLBACK_PLAN_USED", "INITIAL_WEIGHT_REQUIRES_CONFIRMATION"] + reasons
    return GeneratePlanResponse(
        request_id=payload.request_id or str(uuid.uuid4()),
        model_version="fallback-rule-v1",
        feature_schema_version=payload.feature_schema_version,
        split="full_body",
        days=days,
        warnings=warnings,
        fallback=True,
    )


def validate_generated_days(
    days: List[WorkoutDay],
    expected_weekdays: List[int],
    excluded_ids: set[int],
) -> bool:
    """Проверяет структуру дней, порядок weekday и отсутствие запрещённых упражнений."""
    if not isinstance(days, list) or len(days) != len(expected_weekdays):
        return False
    for day, exp_wd in zip(days, expected_weekdays):
        if day.weekday != exp_wd or not day.title or not day.exercises:
            return False
        for ex in day.exercises:
            if ex.exercise_id in excluded_ids:
                return False
            if ex.sets <= 0 or ex.reps_min <= 0 or ex.reps_max < ex.reps_min or ex.rest_sec < 0:
                return False
    return True


def generate_candidate_days(payload: GeneratePlanRequest, weekdays: List[int]) -> List[WorkoutDay]:
    """Формирует план-кандидат с учётом уровня подготовки и исключённых упражнений."""
    excluded = set(payload.profile.excluded_exercise_ids or [])
    level_key = payload.profile.level.strip().lower()

    if "продвин" in level_key or "adv" in level_key:
        sets_delta, min_reps_override, max_reps_override, rest_override = 1, 6, 12, 120
    elif "люб" in level_key or "inter" in level_key:
        sets_delta, min_reps_override, max_reps_override, rest_override = 0, 8, 12, 90
    else:
        sets_delta, min_reps_override, max_reps_override, rest_override = 0, 10, 15, 75

    available = [ex for ex in EXERCISE_CATALOG if ex[0] not in excluded]
    if not available:
        raise ValueError("Все упражнения из каталога попали в excluded_exercise_ids")

    days: List[WorkoutDay] = []
    for idx, day_num in enumerate(weekdays):
        ex1_id, ex1_sets, _, _, ex1_rest = available[idx % len(available)]
        ex2_id, ex2_sets, _, _, ex2_rest = available[(idx + 1) % len(available)]

        items = [
            ExerciseItem(
                exercise_id=ex1_id,
                sets=max(1, ex1_sets + sets_delta),
                reps_min=min_reps_override,
                reps_max=max_reps_override,
                weight_kg=None,
                rest_sec=rest_override if rest_override else ex1_rest,
            )
        ]
        if len(available) > 1 and idx % 2 == 1:
            items.append(
                ExerciseItem(
                    exercise_id=ex2_id,
                    sets=max(1, ex2_sets),
                    reps_min=min_reps_override,
                    reps_max=max_reps_override,
                    weight_kg=None,
                    rest_sec=rest_override if rest_override else ex2_rest,
                )
            )

        days.append(
            WorkoutDay(
                weekday=day_num,
                title=f"День {chr(65 + (idx % 26))}",
                exercises=items,
            )
        )
    return days


@app.get("/health")
def health_check():
    return {"ok": True}


@app.post("/plan/user")
async def plan_user(user: Dict[str, Any] = Body(default_factory=dict)) -> Dict[str, Any]:
    if user:
        logger.info("POST /plan/user keys: %s", list(user.keys())[:12])
    split_name, days = predict(user)
    program = get_random_program(programs=programs, split_name=split_name, days=days)
    if program is None:
        raise HTTPException(status_code=404, detail="Suitable workout program not found")
    return program


@app.post("/v2/plans/generate-local", response_model=GeneratePlanResponse)
def generate_plan_local(payload: GeneratePlanRequest):
    raw_days = payload.profile.training_days
    if not raw_days:
        raise HTTPException(status_code=400, detail="Training days cannot be empty")

    clean_weekdays = [int(d) for d in raw_days if isinstance(d, int) and 0 <= d <= 7]
    req_id = payload.request_id or str(uuid.uuid4())
    excluded = set(payload.profile.excluded_exercise_ids or [])

    if len(clean_weekdays) != len(raw_days):
        fallback_days = clean_weekdays if clean_weekdays else [1, 3, 5]
        return build_fallback_plan(payload, fallback_days, ["INVALID_WEEKDAY_VALUES"])

    try:
        split_name = predict_split_label(payload)
        candidate_days = generate_candidate_days(payload, clean_weekdays)
        if not validate_generated_days(candidate_days, clean_weekdays, excluded):
            logger.warning("Сгенерированный план не прошёл проверку структуры/ограничений; включаем fallback.")
            return build_fallback_plan(payload, clean_weekdays, ["SCHEMA_OR_CONSTRAINT_CHECK_FAILED"])

        model_ver = "catboost+candidate-v2" if _catboost_model is not None else "rule+candidate-v2"
        return GeneratePlanResponse(
            request_id=req_id,
            model_version=model_ver,
            feature_schema_version=payload.feature_schema_version,
            split=split_name,
            days=candidate_days,
            warnings=["INITIAL_WEIGHT_REQUIRES_CONFIRMATION"],
            fallback=False,
        )
    except Exception as exc:
        logger.exception("Ошибка при генерации плана (%s); отдаём безопасный fallback.", exc)
        return build_fallback_plan(payload, clean_weekdays, ["MODEL_GENERATION_EXCEPTION"])


app.include_router(create_router(predict, programs))
