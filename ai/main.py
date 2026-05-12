from typing import Any, Dict
from catboost import CatBoostClassifier
from fastapi import Body, FastAPI
import pandas as pd
import json
import random


programs = []
with open("ai/dataset/dataset.jsonlines", 'r', encoding='utf-8') as f:
    for line in f:
        programs.append(json.loads(line)["output"])



def get_random_program(programs:list, split_name: str, days: list):
    """Функция рандомно выдающая программу тренировок из всего датасета по предсказанному сплиту"""
    filtered = [p for p in programs if p["тип_сплита"] == split_name and len(p['еженедельный_план']) == len(days)]

    if not filtered:
        return None
    
    return random.choice(filtered)



app = FastAPI(title="sportik-ml", version="0.1.0")

# Статический план (совместим с сидом exercises и replaceExercises в Go)
reverse_map = {
    0: "Фулбади",
    1: "ВерхНиз",
    2: "Тяни-толкай",
    3: "Пуш-пул-ноги",
    4: "Сплит"
}


model = CatBoostClassifier()
model.load_model("ai/models/model/workout_model.cbm")
def predict(data):
    days = data["дни_тренировок"]
    df = pd.DataFrame([data])

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
