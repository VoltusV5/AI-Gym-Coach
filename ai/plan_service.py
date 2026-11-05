"""Validated v2 plan assembly around the existing CatBoost split classifier."""

import hmac
import os
from datetime import datetime
from typing import Annotated, Callable, Literal

from fastapi import APIRouter, Depends, HTTPException
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel, ConfigDict, Field, StrictBool, StrictInt, StrictStr, model_validator


PositiveInt = Annotated[StrictInt, Field(gt=0, le=2**31 - 1)]
Text = Annotated[StrictStr, Field(min_length=1, max_length=200, pattern=r"^[^\x00]+$")]
GOALS = {
    "muscle_gain": "Набрать мышцы",
    "weight_loss": "Скинуть вес",
    "fat_loss": "Сжечь жир",
    "recomposition": "Набрать мышцы и сжечь жир",
    "Набрать мышцы": "Набрать мышцы",
    "Сбросить вес": "Скинуть вес",
    "Скинуть вес": "Скинуть вес",
    "Сжечь жир": "Сжечь жир",
    "Набрать мышцы и сжечь жир": "Набрать мышцы и сжечь жир",
}
SPLITS = {"Фулбади": "full_body", "ВерхНиз": "upper_lower", "Тяни-толкай": "push_pull",
          "Пуш-пул-ноги": "push_pull_legs", "Сплит": "split"}
GROUPS = ("Ноги", "Грудь", "Спина", "Плечи", "Руки", "Пресс")
SUBGROUPS = {"Верх спины": "тяга перед собой широким хватом",
             "Широчайшие спины": "тяга сверху узким хватом"}


class DTO(BaseModel):
    model_config = ConfigDict(extra="forbid", allow_inf_nan=False)


class Profile(DTO):
    age: PositiveInt
    height_cm: Annotated[float, Field(strict=True, gt=0, le=2**31 - 1)]
    weight_kg: Annotated[float, Field(strict=True, gt=0, le=2**31 - 1)]
    level: Literal["beginner", "intermediate", "advanced"]
    goal: Text
    training_days: Annotated[list[Annotated[StrictInt, Field(ge=1, le=7)]], Field(min_length=1, max_length=7)]
    equipment_ids: list[Text]
    excluded_exercise_ids: list[PositiveInt] = Field(default_factory=list)
    session_duration_min: PositiveInt = 60
    gender: Text | None = None
    activity_level: Text | None = None
    injuries_notes: StrictBool | None = None

    @model_validator(mode="after")
    def check_categories_and_duplicates(self):
        if self.goal not in GOALS:
            raise ValueError("Unsupported goal")
        for values in (self.training_days, self.equipment_ids, self.excluded_exercise_ids):
            if len(values) != len(set(values)):
                raise ValueError("Duplicate profile constraint")
        return self


class HistoryMetric(DTO):
    exercise_id: PositiveInt
    weight_kg: Annotated[float, Field(strict=True, ge=0)]
    reps: PositiveInt
    volume_kg: Annotated[float, Field(strict=True, ge=0)]
    measured_at: datetime


class History(DTO):
    completed_28d: Annotated[StrictInt, Field(ge=0)] = 0
    adherence_28d: Annotated[float, Field(strict=True, ge=0, le=1)] | None = None
    exercise_metrics: list[HistoryMetric] = Field(default_factory=list)


class ExerciseCandidate(DTO):
    exercise_id: PositiveInt
    muscle_group: Text
    muscle_subgroup: Text | None = None
    equipment_ids: list[Text]


class GeneratePlanRequest(DTO):
    request_id: Text
    feature_schema_version: Literal["profile-v2"]
    profile: Profile
    history: History | None = None
    exercises: Annotated[list[ExerciseCandidate], Field(min_length=1)]

    @model_validator(mode="after")
    def unique_catalog(self):
        ids = [item.exercise_id for item in self.exercises]
        if len(ids) != len(set(ids)):
            raise ValueError("Duplicate catalog exercise")
        return self


class ExerciseItem(DTO):
    exercise_id: PositiveInt
    position: Annotated[StrictInt, Field(ge=0)]
    sets: PositiveInt
    reps_min: PositiveInt
    reps_max: PositiveInt
    weight_kg: Annotated[float, Field(strict=True, ge=0)] | None = None
    rest_sec: Annotated[StrictInt, Field(ge=0)]

    @model_validator(mode="after")
    def ordered_reps(self):
        if self.reps_min > self.reps_max:
            raise ValueError("Invalid repetition range")
        return self


class WorkoutDay(DTO):
    weekday: Annotated[StrictInt, Field(ge=1, le=7)]
    title: Text
    exercises: Annotated[list[ExerciseItem], Field(min_length=1)]


class GeneratePlanResponse(DTO):
    request_id: Text
    model_version: Text
    feature_schema_version: Literal["profile-v2"]
    split: Literal["full_body", "upper_lower", "push_pull", "push_pull_legs", "split"]
    days: Annotated[list[WorkoutDay], Field(min_length=1, max_length=7)]
    warnings: list[Text] = Field(default_factory=list)
    fallback: StrictBool = False


def eligible_exercises(payload: GeneratePlanRequest) -> list[ExerciseCandidate]:
    """Filter independently of the Go caller; no unknown or excluded ID can be selected."""
    available = set(payload.profile.equipment_ids)
    excluded = set(payload.profile.excluded_exercise_ids)
    return [ex for ex in payload.exercises
            if ex.exercise_id not in excluded and set(ex.equipment_ids) <= available]


def duration_seconds(exercises: list[ExerciseItem]) -> int:
    """Conservative scheduling estimate: five minutes setup, three seconds per rep, rest per set."""
    return 300 + sum(ex.sets * (ex.reps_max * 3 + ex.rest_sec) for ex in exercises)


def validate_plan(payload: GeneratePlanRequest, result: GeneratePlanResponse) -> None:
    """Validate the complete candidate against the profile before sending it to Go."""
    allowed = {ex.exercise_id for ex in eligible_exercises(payload)}
    if result.request_id != payload.request_id or result.feature_schema_version != payload.feature_schema_version:
        raise ValueError("Mismatched request")
    weekdays = [day.weekday for day in result.days]
    if sorted(weekdays) != sorted(payload.profile.training_days):
        raise ValueError("Mismatched training days")
    for day in result.days:
        ids = [ex.exercise_id for ex in day.exercises]
        if len(ids) != len(set(ids)) or not set(ids) <= allowed:
            raise ValueError("Invalid exercise selection")
        if [ex.position for ex in day.exercises] != list(range(len(day.exercises))):
            raise ValueError("Invalid exercise order")
        if duration_seconds(day.exercises) > payload.profile.session_duration_min * 60:
            raise ValueError("Session duration exceeded")


def baseline_features(profile: Profile) -> dict | None:
    """Map only explicit supported legacy categories; incomplete profiles use a labelled fallback."""
    gender = {"male": "м", "female": "ж", "м": "м", "ж": "ж"}.get(profile.gender)
    activities = {"Сидячий и малоподвижный", "Лёгкая активность", "Средняя активность", "Высокая активность"}
    if gender is None or profile.activity_level not in activities or profile.injuries_notes is None:
        return None
    return {
        "age": profile.age, "height_cm": profile.height_cm, "weight_kg": profile.weight_kg,
        "gender": gender, "activity_level": profile.activity_level, "injuries_notes": profile.injuries_notes,
        "goal": GOALS[profile.goal],
        "fitness_level": {"beginner": "новичок", "intermediate": "любитель", "advanced": "продвинутый"}[profile.level],
        "training_days_map": profile.training_days,
    }


def assemble_day(slots: list[dict], candidates: list[ExerciseCandidate], weekday: int,
                 index: int, profile: Profile, known: set[int]) -> WorkoutDay:
    """Resolve template muscle slots to current IDs, preferring familiar exercises and varied groups."""
    # Include each major group before additional subgroup slots when time is limited.
    first, remaining, seen = [], [], set()
    for slot in slots:
        group = slot["группа"]
        (remaining if group in seen else first).append(slot)
        seen.add(group)
    required = seen & {"Ноги", "Грудь", "Спина"}
    if not required and first:
        required = {first[0]["группа"]}
    first.sort(key=lambda slot: slot["группа"] not in required)
    selected = []
    used = set()
    covered = set()
    for slot in first + remaining:
        choices = [ex for ex in candidates if ex.muscle_group == slot["группа"] and ex.exercise_id not in used]
        if not choices:
            if slot["группа"] in required - covered:
                raise ValueError("Required muscle group unavailable")
            continue
        subgroup = SUBGROUPS.get(slot.get("подгруппа"), slot.get("подгруппа"))
        exact = [ex for ex in choices if ex.muscle_subgroup == subgroup]
        choices = sorted(exact or choices, key=lambda ex: (ex.exercise_id not in known, ex.exercise_id))
        exercise = choices[index % len(choices)]
        item = ExerciseItem(exercise_id=exercise.exercise_id, position=len(selected),
                            sets=2 if profile.level == "beginner" else 3,
                            reps_min=8, reps_max=12, weight_kg=None, rest_sec=90)
        if duration_seconds(selected + [item]) > profile.session_duration_min * 60:
            if required - covered:
                raise ValueError("Not enough time for required groups")
            break
        selected.append(item)
        used.add(exercise.exercise_id)
        covered.add(exercise.muscle_group)
    return WorkoutDay(weekday=weekday, title=f"День {chr(65 + index)}", exercises=selected)


def generate(payload: GeneratePlanRequest, predict: Callable, templates: dict) -> GeneratePlanResponse:
    """Try the model's compatible template, then a transparent full-body rules fallback."""
    candidates = eligible_exercises(payload)
    known = {m.exercise_id for m in payload.history.exercise_metrics} if payload.history else set()
    warnings = ["INITIAL_WEIGHT_REQUIRES_CONFIRMATION"]
    features = baseline_features(payload.profile)
    split = None
    if features is None:
        warnings.append("BASELINE_FEATURES_UNAVAILABLE")
    else:
        try:
            split, _ = predict(features)
        except Exception:
            # Inference failure never bypasses constraints or exposes model internals.
            warnings.append("BASELINE_INFERENCE_UNAVAILABLE")
    options = [(split, template, False) for template in templates.get((split, len(payload.profile.training_days)), [])]
    fallback_slots = [{"группа": group} for group in GROUPS
                      if any(ex.muscle_group == group for ex in candidates)]
    if {"Ноги", "Грудь", "Спина"} <= {slot["группа"] for slot in fallback_slots}:
        options.append(("Фулбади", [{"упражнения": fallback_slots}] * len(payload.profile.training_days), True))
    for chosen_split, template, fallback in options:
        try:
            days = [assemble_day(day["упражнения"], candidates, weekday, index, payload.profile, known)
                    for index, (weekday, day) in enumerate(zip(payload.profile.training_days, template))]
            result = GeneratePlanResponse(
                request_id=payload.request_id, feature_schema_version=payload.feature_schema_version,
                model_version="rules-v1" if fallback else "catboost-v0-planner-v2",
                split=SPLITS[chosen_split], days=days,
                warnings=warnings + (["RULES_FALLBACK"] if fallback else []), fallback=fallback,
            )
            validate_plan(payload, result)
            return result
        except (ValueError, KeyError):
            continue
    raise HTTPException(status_code=422, detail={"code": "NO_FEASIBLE_PLAN"})


def create_router(predict: Callable, programs: list[dict]) -> APIRouter:
    """Register the private v2 endpoint; preserve the existing v1 app and classifier."""
    router = APIRouter()
    bearer = HTTPBearer(auto_error=False)
    templates = {}
    for program in programs:
        key = (program["тип_сплита"], len(program["еженедельный_план"]))
        days = program["еженедельный_план"]
        if days not in templates.setdefault(key, []):
            templates[key].append(days)

    def authorize(credentials: Annotated[HTTPAuthorizationCredentials | None, Depends(bearer)]):
        expected = os.environ.get("ML_SERVICE_TOKEN", "")
        if not expected:
            raise HTTPException(status_code=503, detail={"code": "ML_UNAVAILABLE"})
        if credentials is None or not hmac.compare_digest(credentials.credentials.encode(), expected.encode()):
            raise HTTPException(status_code=401, detail={"code": "UNAUTHORIZED"})

    @router.post("/v2/plans/generate", response_model=GeneratePlanResponse, dependencies=[Depends(authorize)])
    def generate_plan(payload: GeneratePlanRequest):
        return generate(payload, predict, templates)

    return router
