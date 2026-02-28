import random
import json
from typing import Dict, List, Any

# ========== БАЗА ЗНАНИЙ ==========

EXERCISES = {
    "грудь": {
        "базовые": [
            {"name": "жим штанги лёжа", "вариации": ["жим гантелей", "жим в Смите"]},
            {"name": "жим гантелей лёжа", "вариации": ["жим штанги", "жим под углом"]},
            {"name": "отжимания на брусьях", "вариации": ["жим книзу", "отжимания от скамьи"]}
        ],
        "изолирующие": [
            {"name": "разводка гантелей лёжа", "вариации": ["кроссовер", "разводка в наклоне"]},
            {"name": "кроссовер", "вариации": ["разводка гантелей", "сведение в тренажёре"]}
        ]
    },
    "спина": {
        "базовые": [
            {"name": "подтягивания", "вариации": ["тяга верхнего блока", "австралийские"]},
            {"name": "тяга штанги в наклоне", "вариации": ["тяга гантели", "Т-тяга"]},
            {"name": "тяга гантели к поясу", "вариации": ["тяга штанги", "тяга в тренажёре"]}
        ],
        "изолирующие": [
            {"name": "тяга верхнего блока", "вариации": ["подтягивания", "пулловер"]},
            {"name": "пулловер", "вариации": ["тяга блока", "разгибания"]}
        ]
    },
    "ноги": {
        "базовые": [
            {"name": "приседания со штангой", "вариации": ["фронтальные", "гоблет"]},
            {"name": "жим ногами", "вариации": ["приседания", "гакк-присед"]},
            {"name": "румынская тяга", "вариации": ["мертвая тяга", "на прямых ногах"]}
        ],
        "изолирующие": [
            {"name": "разгибания ног", "вариации": ["приседания", "выпады"]},
            {"name": "сгибания ног", "вариации": ["румынская тяга", "мертвая"]},
            {"name": "выпады с гантелями", "вариации": ["выпады назад", "болгарские"]}
        ]
    },
    "плечи": {
        "базовые": [
            {"name": "жим гантелей сидя", "вариации": ["армейский жим", "жим штанги"]},
            {"name": "армейский жим", "вариации": ["жим гантелей", "жим в Смите"]}
        ],
        "изолирующие": [
            {"name": "махи гантелями в стороны", "вариации": ["кроссовер", "задняя дельта"]},
            {"name": "махи в наклоне", "вариации": ["обратные разведения", "кроссовер"]},
            {"name": "подъёмы гантелей перед собой", "вариации": ["блин", "кроссовер"]}
        ]
    },
    "бицепс": {
        "изолирующие": [
            {"name": "сгибания штанги стоя", "вариации": ["молот", "нижний блок"]},
            {"name": "молотки с гантелями", "вариации": ["сгибания штаги", "концентрированные"]},
            {"name": "сгибания на скамье Скотта", "вариации": ["стоя", "блок"]}
        ]
    },
    "трицепс": {
        "изолирующие": [
            {"name": "французский жим", "вариации": ["разгибания блока", "кикбек"]},
            {"name": "разгибания на блоке", "вариации": ["французский", "обратные отжимания"]}
        ]
    },
    "пресс": {
        "изолирующие": [
            {"name": "скручивания", "вариации": ["кранчи", "с роликом"]},
            {"name": "подъёмы ног", "вариации": ["вис", "обратные скручивания"]},
            {"name": "планка", "вариации": ["боковая", "динамическая"]}
        ]
    }
}

# Сплиты для разного количества дней
SPLITS_BY_DAYS = {
    2: {
        "День 1": ["грудь", "трицепс"],
        "День 2": ["спина", "бицепс"]
    },
    3: {
        "День 1": ["грудь", "трицепс"],
        "День 2": ["спина", "бицепс"],
        "День 3": ["ноги"]
    },
    4: {
        "День 1": ["грудь", "трицепс"],
        "День 2": ["спина", "бицепс"],
        "День 3": ["ноги"],
        "День 4": ["плечи", "пресс"]
    }
}

MOVEMENT_MAP = {
    "жим штанги лёжа": "horizontal_push",
    "жим гантелей лёжа": "horizontal_push",
    "отжимания на брусьях": "vertical_push",

    "жим гантелей сидя": "vertical_push",
    "армейский жим": "vertical_push",

    "подтягивания": "vertical_pull",
    "тяга верхнего блока": "vertical_pull",

    "тяга штанги в наклоне": "horizontal_pull",
    "тяга гантели к поясу": "horizontal_pull",

    "приседания со штангой": "knee_dominant",
    "жим ногами": "knee_dominant",

    "румынская тяга": "hip_dominant",

    "разгибания ног": "knee_isolation",
    "сгибания ног": "hip_isolation",

    "планка": "core",
    "скручивания": "core",
}


EXERCISE_ROLE = {
    "жим штанги лёжа": "base",
    "жим гантелей лёжа": "secondary",
    "разводка гантелей лёжа": "isolation",
    "кроссовер": "isolation",

    "подтягивания": "base",
    "тяга штанги в наклоне": "secondary",
    "тяга верхнего блока": "isolation",

    "приседания со штангой": "base",
    "жим ногами": "secondary",
    "разгибания ног": "isolation",

    "жим гантелей сидя": "base",
    "махи гантелями в стороны": "isolation",

    "сгибания штанги стоя": "secondary",
    "молотки с гантелями": "secondary",

    "французский жим": "secondary",
    "разгибания на блоке": "isolation",

    "скручивания": "secondary",
    "планка": "isolation"
}

MOVEMENT_TYPE = {
    "жим штанги лёжа": "push",
    "жим гантелей лёжа": "push",
    "жим гантелей сидя": "push",

    "подтягивания": "pull",
    "тяга штанги в наклоне": "pull",
    "тяга верхнего блока": "pull"
}

# ========== ПРАВИЛА ПОДБОРА ==========

def pick_structured_exercises(muscle: str, травмы: bool):
    base, secondary, isolation = [], [], []

    for ex_type in EXERCISES.get(muscle, {}):
        for ex in EXERCISES[muscle][ex_type]:
            role = EXERCISE_ROLE.get(ex["name"], "isolation")

            if role == "base":
                base.append(ex)
            elif role == "secondary":
                secondary.append(ex)
            else:
                isolation.append(ex)

    chosen = []

    if not травмы and base:
        chosen.append(random.choice(base))

    if secondary:
        chosen.append(random.choice(secondary))

    if isolation:
        chosen.append(random.choice(isolation))

    return chosen

def balance_movements(exercises: List[Dict]) -> List[Dict]:
    movement_count = {}

    for ex in exercises:
        move = MOVEMENT_MAP.get(ex["основное"])
        if move:
            movement_count[move] = movement_count.get(move, 0) + 1

    # баланс push/pull
    push = movement_count.get("horizontal_push", 0) + movement_count.get("vertical_push", 0)
    pull = movement_count.get("horizontal_pull", 0) + movement_count.get("vertical_pull", 0)

    if push > pull:
        exercises.append({
            "основное": "тяга верхнего блока",
            "вариации": ["подтягивания"]
        })

    # баланс ног
    knee = movement_count.get("knee_dominant", 0)
    hip = movement_count.get("hip_dominant", 0)

    if knee > hip:
        exercises.append({
            "основное": "румынская тяга",
            "вариации": ["на прямых ногах"]
        })

    return exercises

def adjust_by_level(exercises, level):
    if level == "новичок":
        return exercises[:2]
    elif level == "любитель":
        return exercises[:3]
    return exercises

def get_sets_count(level, травмы):
    return 3 if травмы else (3 if level != "продвинутый" else 4)

def get_reps_range(goal):
    return {
        "Сжечь жир": "12-15",
        "Скинуть вес": "12-15",
        "Набрать мышцы": "6-12"
    }.get(goal, "8-12")

def get_rest(level):
    return {
        "новичок": "2 мин",
        "любитель": "1.5 мин",
        "продвинутый": "1 мин"
    }.get(level, "1.5 мин")

def generate_workout(user_input: Dict) -> Dict:

    пол = user_input.get("пол", "м")
    цель = user_input.get("цель", "Набрать мышцы")
    уровень = user_input.get("уровень_подготовки", "новичок")
    травмы = user_input.get("травмы_или_болезни", "нет") == "да"

    дни = user_input.get("дни_тренировок", [])
    кол_дней = max(2, min(4, len(дни))) if дни else 3
    сплит = SPLITS_BY_DAYS.get(кол_дней, SPLITS_BY_DAYS[3])

    план = {}

    for day_name, muscle_groups in сплит.items():

        day_plan = []
        day_exercises_all = []

        # 1️⃣ собираем упражнения по группам
        for muscle in muscle_groups:

            exercises = []

            structured = pick_structured_exercises(muscle, травмы)

            for ex in structured:
                exercises.append({
                    "основное": ex["name"],
                    "вариации": ex["вариации"][:2]
                })

            exercises = adjust_by_level(exercises, уровень)

            day_exercises_all.extend(exercises)

            day_plan.append({
                "группа": muscle,
                "упражнения": exercises,
                "подходы": get_sets_count(уровень, травмы),
                "повторения": get_reps_range(цель),
                "отдых": get_rest(уровень)
            })

        # 2️⃣ балансируем ВЕСЬ день
        balanced = balance_movements(day_exercises_all)

        # 3️⃣ если баланс добавил упражнения — добавляем их в последнюю группу
        if len(balanced) > len(day_exercises_all):
            added = balanced[len(day_exercises_all):]
            day_plan[-1]["упражнения"].extend(added)

        план[day_name] = day_plan

    return {
        "план_тренировок": план,
        "начальные_веса": {}
    }

# ========== ТЕСТ ==========

if __name__ == "__main__":
    with open("ai/generateData/users.json", "r", encoding="utf-8") as file:
        users = json.load(file)

    results = []

    for user in users:
        workout = generate_workout(user)
        
        results.append({
            "user": user,
            "workout": workout
        })

    # сохраняем всё в json
    with open("ai/generateData/generated_workouts.json", "w", encoding="utf-8") as file:
        json.dump(results, file, indent=2, ensure_ascii=False)

    print("Готово ✅ Сохранено в generated_workouts.json")