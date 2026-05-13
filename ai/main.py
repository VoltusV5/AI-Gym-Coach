from typing import Any, Dict
from catboost import CatBoostClassifier
from fastapi import Body, FastAPI
import pandas as pd
import json
import random


programs = []
with open("dataset/dataset.jsonlines", 'r', encoding='utf-8') as f:
    for line in f:
        programs.append(json.loads(line)["output"])



def get_random_program(programs:list, split_name: str, days: list):
    """Функция рандомно выдающая программу тренировок из всего датасета по предсказанному сплиту"""
    filtered = [p for p in programs if p["тип_сплита"] == split_name and len(p['еженедельный_план']) == len(days)]

    if not filtered:
        return None

    return random.choice(filtered)



app = FastAPI(title="sportik-ml", version="0.1.0")


reverse_map = {
    0: "Фулбади",
    1: "ВерхНиз",
    2: "Тяни-толкай",
    3: "Пуш-пул-ноги",
    4: "Сплит"
}


model = CatBoostClassifier()
model.load_model("models/model/workout_model.cbm")
def predict(data):


    mapped = {
        "возраст": data.get("age", data.get("возраст", 30)),
        "рост": data.get("height_cm", data.get("рост", 175)),
        "вес": data.get("weight_kg", data.get("вес", 75)),
    }


    gender = str(data.get("gender", data.get("пол", "м"))).lower()
    if any(x in gender for x in ["жен", "female", "ж"]):
        mapped["пол"] = "ж"
    else:
        mapped["пол"] = "м"



    activity = data.get("activity_level", data.get("тип_активности", "Средняя активность"))
    if "(" in activity:
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

    df['пол_num'] = df['пол'].map({'м': 0, 'ж': 1})
    df['травмы_num'] = df['травмы_или_болезни'].map({'нет': 0, 'да': 1})
    df['уровень_num'] = df['уровень_подготовки'].map({'новичок': 0, 'любитель': 1, 'продвинутый': 2})

    X = df[["возраст", "рост", "вес", "пол_num", "тип_активности", "травмы_num", "цель", "уровень_num"]]

    pred = model.predict(X)[0][0]

    return reverse_map[pred], days


@app.get("/health")
async def health():
    return {"ok": True}


@app.post("/plan/user")
async def plan_user(user: Dict[str, Any] = Body(default_factory=dict)) -> Dict[str, Any]:
    if user:
        print("[ml] POST /plan/user keys:", list(user.keys())[:12])
    split_name, days = predict(user)

    return get_random_program(programs=programs, split_name=split_name, days=days)