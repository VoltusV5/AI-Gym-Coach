import json
from datetime import datetime
from catboost import CatBoostClassifier
import pandas as pd
import random

# Загружаем модель
model = CatBoostClassifier()
model.load_model("ai/models/model/workout_model.cbm")
print("Модель загружена успешно.\n")

EXERCISES_POOL = {
    "Грудь": ["грудные верх", "грудные середина", "грудные низ"],
    "Спина": ["тяга сверху широким хватом", "тяга сверху узким хватом", 
              "тяга перед собой широким хватом", "тяга перед собой узким хватом"],
    "Ноги": ["квадрицепс", "бицепс бедра", "ягодицы", "икры", "приводящие"],
    "Плечи": ["передняя дельта", "средняя дельта", "задняя дельта"],
    "Руки": ["бицепс", "трицепс"],
    "Пресс": [None]
}

def create_day_plan(day_letter: str, day_type: str, num_exercises: int = 7):
    exercises = []
    groups = list(EXERCISES_POOL.keys())
    
    for i in range(num_exercises):
        group = groups[i % len(groups)]
        subgroup_options = EXERCISES_POOL[group]
        subgroup = random.choice(subgroup_options) if subgroup_options and subgroup_options[0] is not None else None
        exercises.append({"группа": group, "подгруппа": subgroup})
    
    return {
        "день": day_letter,
        "тип_дня": day_type,
        "упражнения": exercises
    }

all_plans = []


for i in range(1500):
    profile = {
        "возраст": random.randint(18, 60),
        "рост": random.randint(150, 200),
        "вес": random.randint(45, 120),
        "пол_num": random.choice([0, 1]),
        "тип_активности": random.choice([
            "Сидячий и малоподвижный", 
            "Лёгкая активность", 
            "Средняя активность", 
            "Высокая активность"
        ]),
        "травмы_num": random.choice([0, 1]),
        "цель": random.choice([
            "Набрать мышцы", 
            "Сжечь жир", 
            "Скинуть вес", 
            "Набрать мышцы и сжечь жир"
        ]),
        "уровень_num": random.choice([0, 1, 2])
    }
    
    input_df = pd.DataFrame([profile])
    split_num = int(model.predict(input_df)[0][0])
    
    split_names = ["Фулбади", "ВерхНиз", "Тяни-толкай", "Пуш-пул-ноги", "Сплит"]
    split_name = split_names[split_num % len(split_names)]
    
    # Структура дней
    if split_name in ["Фулбади", "ВерхНиз"]:
        day_letters = ["A", "B", "C"]
    else:
        day_letters = ["A", "B", "C", "D"]
    
    weekly_plan = []
    for letter in day_letters:
        day_type = f"{split_name} {letter}"
        num_ex = 8 if split_name == "Фулбади" else 6
        day_plan = create_day_plan(letter, day_type, num_exercises=num_ex)
        weekly_plan.append(day_plan)
    
    plan = {
        "тип_сплита": split_name,
        "еженедельный_план": weekly_plan
    }
    
    all_plans.append(plan)
    
    if (i + 1) % 20 == 0:
        print(f"Создано {i+1:3d} планов")


timestamp = datetime.now().strftime("%Y-%m-%d_%H-%M")
filename = f"many_workout_plans_200_{timestamp}.json"

final_output = {
    "total_plans": len(all_plans),
    "generated_at": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
    "plans": all_plans
}

with open(filename, "w", encoding="utf-8") as f:
    json.dump(final_output, f, ensure_ascii=False, indent=2)

print("\n" + "="*70)