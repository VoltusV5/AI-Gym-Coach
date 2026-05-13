
"""
fetch_real_exercise_media_v3.py
-------------------------------
Downloads REAL exercise images (person performing) from:
  - github.com/yuhonas/free-exercise-db  (873 exercises, CC0)
  - wger.de open API for animated GIFs

Correct URL format:
  https://raw.githubusercontent.com/yuhonas/free-exercise-db/main/exercises/{id}/images/0.jpg

Run from fitness-app root:
  python -X utf8 scripts/fetch_real_exercise_media_v3.py
"""

import os, sys, io, json, time, shutil, urllib.request, urllib.parse

if sys.stdout.encoding != "utf-8":
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8", errors="replace")

STATIC_DIR = os.path.normpath(
    os.path.join(os.path.dirname(__file__), "..", "backend", "static", "exercises")
)
os.makedirs(STATIC_DIR, exist_ok=True)

RAW_GH  = "https://raw.githubusercontent.com/yuhonas/free-exercise-db/main"
IMG_BASE = f"{RAW_GH}/exercises"

HEADERS = {
    "User-Agent": ("Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
                   "AppleWebKit/537.36 (KHTML, like Gecko) "
                   "Chrome/124.0 Safari/537.36"),
    "Accept": "*/*",
}





MANUAL_MAP = {

    "Жим ногами лёжа":                                          "Leg_Press",
    "Приседания со штангой":                                    "Barbell_Squat",
    "Приседания в гаке":                                        "Hack_Squat",
    "Разгибания ног в тренажёре сидя":                          "Leg_Extensions",
    "Болгарские выпады с гантелями":                            "Barbell_Side_Split_Squat",

    "Сгибания ног в тренажёре сидя":                            "Seated_Leg_Curl",
    "Сгибания ног в тренажёре лёжа":                            "Lying_Leg_Curls",
    "Румынская тяга со штангой":                                "Romanian_Deadlift",
    "Румынская тяга с гантелями":                               "Romanian_Deadlift",

    "Гиперэкстензия":                                           "Hyperextensions_Back_Extensions",
    "Ягодичный мост в тренажёре":                               "Barbell_Hip_Thrust",
    "Ягодичный мост со штангой":                                "Barbell_Hip_Thrust",
    "Разведения ног в стороны в тренажёре сидя":                "Hip_Abductor_Exercise_Machine",
    "Сведения ног друг к другу сидя в тренажёре":               "Hip_Adduction_Machine",
    "Отведение ног назад в кроссовере":                         "Donkey_Kickbacks",
    "Отведения ног назад в тренажёре":                          "Donkey_Kickbacks",

    "Подъёмы на носки стоя в тренажёре":                        "Standing_Calf_Raises",
    "Подъёмы на носки сидя в тренажёре":                        "Seated_Calf_Raise",
    "Подъёмы на носки стоя со штангой":                         "Standing_Calf_Raises",
    "Подъёмы на носки стоя с гирей или гантелей на одной ноге": "Standing_Calf_Raises",
    "Жим носками в тренажёре для жима ног":                     "Donkey_Calf_Raises",

    "Скручивания в тренажёре":                                   "Ab_Crunch_Machine",
    "Молитва":                                                   "Cable_Crunch",

    "Махи с гантелями в стороны":                               "Dumbbell_Lying_One-Arm_Rear_Lateral_Raise",
    "Отведение руки в сторону в кроссовере":                     "Cable_Seated_Lateral_Raise",
    "Жим гантелей сидя над головой":                            "Dumbbell_Shoulder_Press",
    "Жим сидя на плечи в тренажёре":                            "Machine_Shoulder_(Military)_Press",
    "Махи с гантелями в стороны в наклоне сидя":                "Bent_Over_Dumbbell_Rear_Delt_Raise_With_Head_On_Bench",
    "Махи с гантелями в стороны в наклоне стоя":                "Bent_Over_Dumbbell_Rear_Delt_Raise_With_Head_On_Bench",
    "Бабочка на плечи":                                         "Barbell_Rear_Delt_Row",

    "Разгибания рук в кроссовере":                              "Triceps_Pushdown",
    "Фрунзуский жим с гантелями":                               "Decline_Close-Grip_Bench_To_Skull_Crusher",
    "Французский жим со штангой":                               "EZ-Bar_Skullcrusher",
    "Разгибание рук с гантелью над головой":                    "Dumbbell_Tricep_Extension_-Pronated_Grip",
    "Разгибания рук в кроссовере из-за головы":                 "Cable_One_Arm_Tricep_Extension",

    "Сгибания рук в кроссовере":                                "Cable_Hammer_Curls_-_Rope_Attachment",
    "Сгибания рук со штангой стоя":                             "Barbell_Curl",
    "Сгибания рук с гантелями стоя":                            "Dumbbell_Alternate_Bicep_Curl",
    "Сгибания рук со штангой на скамье Скотта":                 "Preacher_Curl",
    "Сгибания рук с гантелями на скамье Скотта":               "One_Arm_Dumbbell_Preacher_Curl",
    "Молотки":                                                   "Hammer_Curls",
    "Сгибания рук с гантелями сидя":                            "Seated_Dumbbell_Curl",

    "Жим в хамере на верх груди":                               "Incline_Push-Up",
    "Жим в хамере на середину груди":                           "Dumbbell_Bench_Press",
    "Жим в хамере на низ груди":                                "Decline_Dumbbell_Bench_Press",
    "Жим штанги на горизонтальной скамье":                      "Barbell_Bench_Press_-_Medium_Grip",
    "Жим штанги на наклонной скамье":                           "Incline_Barbell_Triceps_Extension",
    "Жим штанги на скамье с обратным наклоном":                 "Decline_Barbell_Bench_Press",
    "Жим гантелей на наклонной скамье":                         "Incline_Dumbbell_Press",
    "Жим гантелей на горизонтальной скамье":                    "Dumbbell_Bench_Press",
    "Жим гантелей на скамье с обратным наклоном":               "Decline_Dumbbell_Bench_Press",
    "Отжимания на брусьях":                                      "Dips_-_Chest_Version",
    "Отжимания на брусьях с отягощением":                        "Weighted_Dips",
    "Сведение рук в кроссовере на верх груди":                   "Cable_Crossover",
    "Сведение рук в кроссовере на середину груди":               "Cable_Crossover",
    "Сведение рук в кроссовере на низ груди":                    "Low_Cable_Crossover",
    "Бабочка на грудь":                                          "Butterfly",

    "Подтягивания":                                              "Wide-Grip_Lat_Pulldown",
    "Подтягивания с отягощением":                                "Weighted_Pull_Ups",
    "Горизонтальная тяга с узкой ручкой":                        "Seated_Cable_Rows",
    "Горизонтальная тяга с широкой ручкой":                      "Seated_Cable_Rows",
    "Вертикальная тяга с узкой ручкой":                          "Close-Grip_Front_Lat_Pulldown",
    "Вертикальная тяга с широкой ручкой":                        "Wide-Grip_Lat_Pulldown",
    "Т-гриф":                                                    "T-Bar_Row_with_Handle",
    "Тяга штанги к поясу":                                       "Bent_Over_Barbell_Row",
    "Тяга гантели к поясу":                                      "One-Arm_Dumbbell_Row",
    "Пуловер":                                                   "Bent-Arm_Dumbbell_Pullover",
}

EXERCISES = [
    (1,  "Жим ногами лёжа"),
    (2,  "Приседания со штангой"),
    (3,  "Приседания в гаке"),
    (4,  "Разгибания ног в тренажёре сидя"),
    (5,  "Болгарские выпады с гантелями"),
    (6,  "Сгибания ног в тренажёре сидя"),
    (7,  "Сгибания ног в тренажёре лёжа"),
    (8,  "Румынская тяга со штангой"),
    (9,  "Румынская тяга с гантелями"),
    (10, "Гиперэкстензия"),
    (11, "Ягодичный мост в тренажёре"),
    (12, "Ягодичный мост со штангой"),
    (13, "Разведения ног в стороны в тренажёре сидя"),
    (14, "Сведения ног друг к другу сидя в тренажёре"),
    (15, "Отведение ног назад в кроссовере"),
    (16, "Отведения ног назад в тренажёре"),
    (17, "Подъёмы на носки стоя в тренажёре"),
    (18, "Подъёмы на носки сидя в тренажёре"),
    (19, "Подъёмы на носки стоя со штангой"),
    (20, "Подъёмы на носки стоя с гирей или гантелей на одной ноге"),
    (21, "Жим носками в тренажёре для жима ног"),
    (22, "Скручивания в тренажёре"),
    (23, "Молитва"),
    (24, "Махи с гантелями в стороны"),
    (25, "Отведение руки в сторону в кроссовере"),
    (26, "Жим гантелей сидя над головой"),
    (27, "Жим сидя на плечи в тренажёре"),
    (28, "Махи с гантелями в стороны в наклоне сидя"),
    (29, "Махи с гантелями в стороны в наклоне стоя"),
    (30, "Бабочка на плечи"),
    (31, "Разгибания рук в кроссовере"),
    (32, "Фрунзуский жим с гантелями"),
    (33, "Французский жим со штангой"),
    (34, "Разгибание рук с гантелью над головой"),
    (35, "Разгибания рук в кроссовере из-за головы"),
    (36, "Сгибания рук в кроссовере"),
    (37, "Сгибания рук со штангой стоя"),
    (38, "Сгибания рук с гантелями стоя"),
    (39, "Сгибания рук со штангой на скамье Скотта"),
    (40, "Сгибания рук с гантелями на скамье Скотта"),
    (41, "Молотки"),
    (42, "Сгибания рук с гантелями сидя"),
    (43, "Жим в хамере на верх груди"),
    (44, "Жим в хамере на середину груди"),
    (45, "Жим в хамере на низ груди"),
    (46, "Жим штанги на горизонтальной скамье"),
    (47, "Жим штанги на наклонной скамье"),
    (48, "Жим штанги на скамье с обратным наклоном"),
    (49, "Жим гантелей на наклонной скамье"),
    (50, "Жим гантелей на горизонтальной скамье"),
    (51, "Жим гантелей на скамье с обратным наклоном"),
    (52, "Отжимания на брусьях"),
    (53, "Отжимания на брусьях с отягощением"),
    (54, "Сведение рук в кроссовере на верх груди"),
    (55, "Сведение рук в кроссовере на середину груди"),
    (56, "Сведение рук в кроссовере на низ груди"),
    (57, "Бабочка на грудь"),
    (58, "Подтягивания"),
    (59, "Подтягивания с отягощением"),
    (60, "Горизонтальная тяга с узкой ручкой"),
    (61, "Горизонтальная тяга с широкой ручкой"),
    (62, "Вертикальная тяга с узкой ручкой"),
    (63, "Вертикальная тяга с широкой ручкой"),
    (64, "Т-гриф"),
    (65, "Тяга штанги к поясу"),
    (66, "Тяга гантели к поясу"),
    (67, "Пуловер"),
]


def get(url: str, timeout: int = 30) -> bytes | None:
    try:
        req = urllib.request.Request(url, headers=HEADERS)
        with urllib.request.urlopen(req, timeout=timeout) as r:
            data = r.read()
        return data if len(data) > 1000 else None
    except Exception as e:
        print(f"    FAIL  {url.split('/')[-1]}  ({type(e).__name__})")
        return None


def try_dl(url: str, dest: str) -> bool:
    if os.path.isfile(dest) and os.path.getsize(dest) > 1000:
        return True
    data = get(url)
    if data:
        with open(dest, "wb") as f:
            f.write(data)
        print(f"    OK   {os.path.basename(dest)}  ({len(data)//1024}KB)")
        return True
    return False


def load_db() -> dict:
    cache = os.path.join(STATIC_DIR, "_exercises_db.json")
    if os.path.isfile(cache):
        with open(cache, encoding="utf-8") as f:
            raw = json.load(f)
        print(f"Loaded {len(raw)} exercises from cache")
    else:
        print("Downloading exercises.json from GitHub...")
        data = get(
            "https://raw.githubusercontent.com/yuhonas/free-exercise-db/main/dist/exercises.json",
            timeout=60
        )
        if not data:
            return {}
        with open(cache, "wb") as f:
            f.write(data)
        raw = json.loads(data)
        print(f"Downloaded {len(raw)} exercises")
    return {ex["id"]: ex for ex in raw}


def load_wger_gifs() -> dict[int, str]:
    cache = os.path.join(STATIC_DIR, "_wger_images.json")
    if os.path.isfile(cache):
        with open(cache, encoding="utf-8") as f:
            imgs = json.load(f)
        print(f"Loaded {len(imgs)} wger images from cache")
    else:
        print("Fetching wger exercise image list...")
        imgs = []
        url = "https://wger.de/api/v2/exerciseimage/?format=json&limit=100&offset=0"
        while url:
            data = get(url, timeout=30)
            if not data:
                break
            page = json.loads(data)
            imgs.extend(page.get("results", []))
            url = page.get("next")
            time.sleep(0.3)
        with open(cache, "w", encoding="utf-8") as f:
            json.dump(imgs, f)
        print(f"Fetched {len(imgs)} wger images")

    gifs = {}
    for img in imgs:
        url = img.get("image", "")
        if url.lower().endswith(".gif"):
            eid = img.get("exercise_base")
            if eid and eid not in gifs:
                gifs[eid] = url
    return gifs


def wger_search(name: str) -> int | None:
    url = f"https://wger.de/api/v2/exercise/search/?term={urllib.parse.quote(name)}&format=json"
    data = get(url, timeout=15)
    if not data:
        return None
    try:
        j = json.loads(data)
        s = j.get("suggestions", [])
        if s:
            return s[0].get("data", {}).get("base_id")
    except Exception:
        pass
    return None


def main():
    print(f"Static dir: {STATIC_DIR}\n")
    db = load_db()
    if not db:
        print("FATAL: could not load exercises DB")
        return

    print("\nLoading wger GIF map...")
    gif_map = load_wger_gifs()
    print(f"  wger GIFs available: {len(gif_map)}\n")

    sql_lines = [
        "-- REAL exercise media URLs (free-exercise-db + wger)",
        "-- Run: psql -U sportapp_user -d sportapp_db -f backend/static/exercises/update_media_urls.sql\n",
    ]

    img_count, gif_count = 0, 0

    for our_id, ru_name in EXERCISES:
        print(f"[{our_id:02d}] {ru_name}")
        img_dest = os.path.join(STATIC_DIR, f"img_{our_id}.jpg")
        gif_dest = os.path.join(STATIC_DIR, f"vid_{our_id}.gif")


        img_ok = os.path.isfile(img_dest) and os.path.getsize(img_dest) > 1000
        gif_ok = os.path.isfile(gif_dest) and os.path.getsize(gif_dest) > 1000


        if not img_ok:
            ex_id = MANUAL_MAP.get(ru_name)
            ex = db.get(ex_id) if ex_id else None
            if ex and ex.get("images"):
                for rel in ex["images"]:

                    parts = rel.split("/")
                    fid, fname = parts[0], parts[-1]
                    url = f"{IMG_BASE}/{fid}/images/{fname}"
                    if try_dl(url, img_dest):
                        img_ok = True
                        img_count += 1
                        break
            if not img_ok:
                print(f"    [img] not found in free-exercise-db")

        else:
            print(f"    img cached")


        if not gif_ok:

            ex_id = MANUAL_MAP.get(ru_name)
            ex = db.get(ex_id) if ex_id else None
            en_name = ex["name"] if ex else ru_name

            wger_id = wger_search(en_name)
            time.sleep(0.25)

            if wger_id and wger_id in gif_map:
                if try_dl(gif_map[wger_id], gif_dest):
                    gif_ok = True
                    gif_count += 1


            if not gif_ok and wger_id:
                info_url = f"https://wger.de/api/v2/exerciseinfo/{wger_id}/?format=json"
                data = get(info_url, timeout=15)
                if data:
                    try:
                        info = json.loads(data)
                        for img in info.get("images", []):
                            img_url = img.get("image", "")
                            if img_url:
                                if try_dl(img_url, gif_dest):
                                    gif_ok = True
                                    gif_count += 1
                                    break
                    except Exception:
                        pass
                time.sleep(0.2)


            if not gif_ok and img_ok:
                shutil.copy2(img_dest, gif_dest)
                print(f"    [vid] fallback: copy of img")
                gif_ok = True
                gif_count += 1

        else:
            print(f"    vid cached")


        image_url = f"/static/exercises/img_{our_id}.jpg" if img_ok else None
        video_url  = f"/static/exercises/vid_{our_id}.gif" if gif_ok else None
        sets = []
        if image_url:
            sets.append(f"image_url = '{image_url}'")
        if video_url:
            sets.append(f"video_url = '{video_url}'")
        if sets:
            sql_lines.append(
                f"UPDATE sportapp.exercises SET {', '.join(sets)} WHERE id = {our_id};"
            )

    sql_path = os.path.join(STATIC_DIR, "update_media_urls.sql")
    with open(sql_path, "w", encoding="utf-8") as f:
        f.write("\n".join(sql_lines) + "\n")

    print(f"\nDone!")
    print(f"  Real JPG images:  {img_count} new  (total {len([f for f in os.listdir(STATIC_DIR) if f.startswith('img_') and f.endswith('.jpg')])})")
    print(f"  GIF / video files:{gif_count} new  (total {len([f for f in os.listdir(STATIC_DIR) if f.startswith('vid_') and f.endswith('.gif')])})")
    print(f"  SQL: {sql_path}")


if __name__ == "__main__":
    main()