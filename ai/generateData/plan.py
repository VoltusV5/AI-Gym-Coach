import json
import random
import jsonlines





MUSCLES = {
    "Шея": [None],
    "Плечи": ["передняя дельта", "средняя дельта", "задняя дельта"],
    "Руки": ["трицепс", "бицепс"],
    "Грудь": ["грудные верх", "грудные середина", "грудные низ"],
    "Спина": ["Широчайшие спины", "Верх спины"],
    "Пресс": [None],
    "Ноги": ["ягодицы", "бицепс бедра", "квадрицепс", "икры"]
}


def add_group(exercises, group):

    subs = MUSCLES[group]

    if subs == [None]:
        exercises.append({"группа": group, "подгруппа": None})
    else:
        for sub in subs:
            exercises.append({"группа": group, "подгруппа": sub})






def choose_sub(group, used):

    options = MUSCLES[group]

    if options == [None]:
        return None






    available = [o for o in options if o not in used]

    if not available:
        available = options.copy()
        used.clear()
    sub = random.choice(available)

    used.add(sub)

    return sub


def ex(group, used):

    return {
        "группа": group,
        "подгруппа": choose_sub(group, used)
    }





def fullbody(label):

    used = set()

    exercises = [
        ex("Грудь", used),
        ex("Грудь", used),
        ex("Спина", used),
        ex("Спина", used),
        ex("Ноги", used),
        ex("Ноги", used),
        ex("Плечи", used),
        ex("Руки", used),
        {"группа": "Пресс", "подгруппа": None}
    ]

    return {
        "день": label,
        "тип_дня": f"Фулбоди {label}",
        "упражнения": exercises
    }





def upper(label):

    used = set()

    exercises = [
        ex("Грудь", used),
        ex("Спина", used),
        ex("Плечи", used),
        ex("Плечи", used),
        ex("Руки", used),
        ex("Руки", used)
    ]

    return {
        "день": label,
        "тип_дня": "Верх",
        "упражнения": exercises
    }





def lower(label):

    used = set()

    exercises = [
        ex("Ноги", used),
        ex("Ноги", used),
        ex("Ноги", used),
        ex("Ноги", used),
        {"группа": "Пресс", "подгруппа": None}
    ]

    return {
        "день": label,
        "тип_дня": "Низ",
        "упражнения": exercises
    }





def push(label):

    used = set()

    exercises = [
        ex("Грудь", used),
        ex("Грудь", used),
        {"группа": "Плечи", "подгруппа": "передняя дельта"},
        {"группа": "Плечи", "подгруппа": "средняя дельта"},
        {"группа": "Руки", "подгруппа": "трицепс"},
        {"группа": "Пресс", "подгруппа": None}
    ]

    return {
        "день": label,
        "тип_дня": "Толкай",
        "упражнения": exercises
    }





def pull(label):

    used = set()

    exercises = [
        ex("Спина", used),
        ex("Спина", used),
        {"группа": "Плечи", "подгруппа": "задняя дельта"},
        {"группа": "Руки", "подгруппа": "бицепс"},
        {"группа": "Пресс", "подгруппа": None}
    ]

    return {
        "день": label,
        "тип_дня": "Тяни",
        "упражнения": exercises
    }





def split_day(label, template):

    exercises = []

    for g in template:

        if g == "бицепс":
            exercises.append({"группа": "Руки", "подгруппа": "бицепс"})

        elif g == "трицепс":
            exercises.append({"группа": "Руки", "подгруппа": "трицепс"})

        elif g == "пресс":
            exercises.append({"группа": "Пресс", "подгруппа": None})

        else:
            add_group(exercises, g)

    return {
        "день": label,
        "тип_дня": "Сплит",
        "упражнения": exercises
    }





def choose_split(user):

    days = len(user["дни_тренировок"])
    level = user["уровень_подготовки"]
    injuries = user["травмы_или_болезни"]

    if injuries == "да":
        return random.choice(["ВерхНиз", "Сплит"])

    if level == "новичок":

        if days <= 3:
            return "Фулбади"

    if level == "любитель":

        if days == 3:
            return random.choice(["Фулбади",  "Тяни-толкай", "Тяни-толкай"])

        if days >= 4:
            return random.choice(["ВерхНиз", "Сплит", "Тяни-толкай"])

        if level == "продвинутый":

            if days >= 4:
                return random.choice([
                    "Тяни-толкай",
                    "ВерхНиз",
                    "ВерхНиз",
                    "Сплит"
                ])

    return random.choice(["Фулбади", "ВерхНиз", "Тяни-толкай"])





def generate_plan(user):

    split = choose_split(user)

    days = len(user["дни_тренировок"])

    week = []

    if split == "Фулбади":

        for i in range(days):
            week.append(fullbody(chr(65+i)))

    elif split == "ВерхНиз":

        funcs = [upper, lower]

        for i in range(days):
            week.append(funcs[i % 2](chr(65+i)))

    elif split == "Тяни-толкай":

        funcs = [push, pull]

        for i in range(days):
            week.append(funcs[i % 2](chr(65+i)))

    elif split == "Сплит":

        template = [
            ["Грудь","трицепс"],
            ["Спина","бицепс"],
            ["Ноги"],
            ["Плечи","пресс"]
        ]

        for i in range(days):
            week.append(split_day(chr(65+i), template[i % len(template)]))

    return {
        "тип_сплита": split,
        "еженедельный_план": week
    }





with open("ai/generateData/users.json", "r",encoding="utf-8") as f:
    users = json.load(f)

dataset = []

for user in users:

    plan = generate_plan(user)

    dataset.append({
        "input": user,
        "output": plan
    })

with jsonlines.open("dataset_with_targets.jsonlines",mode="w") as f:
    f.write_all(dataset)
