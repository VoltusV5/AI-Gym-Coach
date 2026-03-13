import json
import matplotlib.pyplot as plt


# открываем JSONL и читаем построчно
sp = []
with open("ai/generateData/dataset.jsonlines", "r", encoding="utf-8") as f:
    for line in f:
        example = json.loads(line)
        sp.append(example["output"]["тип_сплита"])

# считаем распределение
data = {i: sp.count(i) for i in set(sp)}

# строим график
plt.bar(list(data.keys()), [v / sum(data.values()) * 100 for v in data.values()])
plt.title("Распределение сплитов")
plt.ylabel("Процент (%)")
plt.show()

with open("ai/generateData/dataset_users.json", "r",encoding="utf-8") as f:
    users = json.load(f)

sp = []
for i in users:
    sp.append(i["уровень_подготовки"])


data = {i:(sp.count(i)) for i in sp}

fig, ax = plt.subplots(1, 2)

ax[0].bar(list(data.keys()), [i/sum(data.values())*100 for i in data.values()])


sp = []
for i in users:
    sp.append(i["возраст"])

data = {i:(sp.count(i)) for i in sp}

ax[1].bar(list(data.keys()), list(data.values()))
plt.show()

