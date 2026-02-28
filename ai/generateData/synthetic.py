import json
import pandas as pd
import random

def input_json():
    return {
        "возраст": random.randint(13, 65),
        "рост": 0,
        "вес": 0,
        "пол": random.choice(["м", "ж"]),
        "тип_активности": None,
        "травмы_или_болезни": random.choice(["да", "нет"]),
        "цель": None,
        "уровень_подготовки": None,
        "дни_тренировок": None,
    }


def weighted_choice(options, weights):
    return random.choices(options, weights=weights)[0]


def generate_user():
    users = []

    while len(users) < 2000:
        user = input_json()
        age = user["возраст"]
        injuries = user["травмы_или_болезни"]

        if age < 18:
            goal = weighted_choice(
                ["Набрать мышцы", "Набрать мышцы и сжечь жир", "Сжечь жир"],
                [50, 30, 20]
            )
        elif 18 <= age <= 35:
            goal = weighted_choice(
                ["Набрать мышцы", "Набрать мышцы и сжечь жир", "Сжечь жир", "Скинуть вес"],
                [30, 30, 25, 15]
            )
        else:
            goal = weighted_choice(
                ["Сжечь жир", "Скинуть вес", "Набрать мышцы"],
                [40, 35, 25]
            )

        user["цель"] = goal

        if user["пол"] == "м":
            height = int(random.normalvariate(175, 7))
        else:
            height = int(random.normalvariate(165, 6))

        height = max(150, min(height, 200))
        user["рост"] = height

        if age < 18:
            bmi = random.uniform(17, 23)
        elif goal in ["Скинуть вес", "Сжечь жир"]:
            bmi = random.uniform(24, 32)
        elif goal == "Набрать мышцы":
            bmi = random.uniform(21, 27)
        else:
            bmi = random.uniform(20, 26)

        weight = int(bmi * (height / 100) ** 2)
        user["вес"] = weight

        if 13 <= age <= 15:
            level = weighted_choice(["новичок", "любитель"], [80, 20])
        elif 16 <= age <= 19:
            level = weighted_choice(["новичок", "любитель", "продвинутый"], [60, 30, 10])
        elif 20 <= age <= 29:
            level = weighted_choice(["новичок", "любитель", "продвинутый"], [40, 40, 20])
        elif 30 <= age <= 39:
            level = weighted_choice(["новичок", "любитель", "продвинутый"], [50, 35, 15])
        elif 40 <= age <= 49:
            level = weighted_choice(["новичок", "любитель", "продвинутый"], [60, 30, 10])
        else:
            level = weighted_choice(["новичок", "любитель", "продвинутый"], [75, 20, 5])

        if injuries == "да" and level == "продвинутый":
            level = "любитель"

        user["уровень_подготовки"] = level

        if level == "новичок":
            training_count = weighted_choice([2, 3, 4], [50, 40, 10])
        elif level == "любитель":
            training_count = weighted_choice([2, 3, 4, 5], [10, 40, 35, 15])
        else:
            training_count = weighted_choice([3, 4, 5], [20, 40, 40])

        if injuries == "да":
            training_count = min(training_count, 3)

        days = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"]
        user["дни_тренировок"] = sorted(
            random.sample(days, training_count),
            key=days.index
        )

        if level == "новичок":
            activity = weighted_choice(
                ["Сидячий и малоподвижный", "Лёгкая активность", "Средняя активность"],
                [50, 35, 15]
            )
        elif level == "любитель":
            activity = weighted_choice(
                ["Лёгкая активность", "Средняя активность", "Высокая активность"],
                [30, 40, 30]
            )
        else:
            activity = weighted_choice(
                ["Средняя активность", "Высокая активность"],
                [40, 60]
            )

        if injuries == "да" and activity == "Высокая активность":
            activity = "Средняя активность"

        user["тип_активности"] = activity

        users.append(user)

    return users


with open("users.json", "w") as file:
    json.dump(generate_user(), file, indent=4, ensure_ascii=False)

#Примеры с очень высокой активностью
#Примеры детишек 14 лет и наоборот постарше
#Интересные случаи 


