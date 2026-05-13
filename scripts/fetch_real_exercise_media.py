
"""
fetch_real_exercise_media.py
----------------------------
Downloads REAL exercise images (person performing the exercise) from:
  - Images (JPG/PNG): github.com/yuhonas/free-exercise-db  (CC0 / public domain)
  - Animated GIFs: wger.de open-source API (where available)

Maps every Russian exercise name to the closest English equivalent in the
free-exercise-db dataset using a manual curated override table, then falls
back to a word-overlap fuzzy scorer for any unmatched exercises.

Usage:
  python -X utf8 scripts/fetch_real_exercise_media.py
"""

import os
import sys
import io
import json
import time
import urllib.request
import urllib.parse

if sys.stdout.encoding != "utf-8":
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8", errors="replace")

STATIC_DIR = os.path.normpath(
    os.path.join(os.path.dirname(__file__), "..", "backend", "static", "exercises")
)
os.makedirs(STATIC_DIR, exist_ok=True)




RAW_GH = "https://raw.githubusercontent.com/yuhonas/free-exercise-db/main"
EXERCISES_JSON_URL = f"{RAW_GH}/dist/exercises.json"





MANUAL_MAP = {
    "Жим ногами лёжа":                                          "Leg_Press",
    "Приседания со штангой":                                    "Barbell_Squat",
    "Приседания в гаке":                                        "Hack_Squat",
    "Разгибания ног в тренажёре сидя":                          "Leg_Extensions",
    "Болгарские выпады с гантелями":                            "Dumbbell_Bulgarian_Split_Squat",
    "Сгибания ног в тренажёре сидя":                            "Seated_Leg_Curl",
    "Сгибания ног в тренажёре лёжа":                            "Lying_Leg_Curls",
    "Румынская тяга со штангой":                                "Barbell_Romanian_Deadlift",
    "Румынская тяга с гантелями":                               "Romanian_Deadlift_Dumbbell",
    "Гиперэкстензия":                                           "Hyperextension",
    "Ягодичный мост в тренажёре":                               "Barbell_Hip_Thrust",
    "Ягодичный мост со штангой":                                "Barbell_Hip_Thrust",
    "Разведения ног в стороны в тренажёре сидя":                "Hip_Abductor_Exercise_Machine",
    "Сведения ног друг к другу сидя в тренажёре":               "Hip_Adduction_Machine",
    "Отведение ног назад в кроссовере":                         "Cable_Hip_Adduction",
    "Отведения ног назад в тренажёре":                          "Donkey_Kickbacks",
    "Подъёмы на носки стоя в тренажёре":                        "Standing_Calf_Raises",
    "Подъёмы на носки сидя в тренажёре":                        "Seated_Calf_Raise",
    "Подъёмы на носки стоя со штангой":                         "Standing_Calf_Raises",
    "Подъёмы на носки стоя с гирей или гантелей на одной ноге": "Standing_Calf_Raises",
    "Жим носками в тренажёре для жима ног":                     "Donkey_Calf_Raises",
    "Скручивания в тренажёре":                                   "Machine_Crunch",
    "Молитва":                                                   "Kneeling_Cable_Crunch",
    "Махи с гантелями в стороны":                               "Dumbbell_Lateral_Raise",
    "Отведение руки в сторону в кроссовере":                     "Cable_Lateral_Raise",
    "Жим гантелей сидя над головой":                            "Dumbbell_Shoulder_Press",
    "Жим сидя на плечи в тренажёре":                            "Seated_Bradford_Rocky_Press",
    "Махи с гантелями в стороны в наклоне сидя":                "Dumbbell_Rear_Lateral_Raise",
    "Махи с гантелями в стороны в наклоне стоя":                "Dumbbell_Rear_Lateral_Raise",
    "Бабочка на плечи":                                         "Rear_Delt_Fly",
    "Разгибания рук в кроссовере":                              "Triceps_Pushdown_-_Rope_Attachment",
    "Фрунзуский жим с гантелями":                               "Lying_Dumbbell_Tricep_Extension",
    "Французский жим со штангой":                               "EZ_Bar_Skullcrusher",
    "Разгибание рук с гантелью над головой":                    "Dumbbell_Tricep_Extension",
    "Разгибания рук в кроссовере из-за головы":                 "Cable_Overhead_Triceps_Extension",
    "Сгибания рук в кроссовере":                                "Cable_Hammer_Curl_-_Rope_Attachment",
    "Сгибания рук со штангой стоя":                             "Barbell_Curl",
    "Сгибания рук с гантелями стоя":                            "Dumbbell_Alternate_Bicep_Curl",
    "Сгибания рук со штангой на скамье Скотта":                 "Preacher_Curl",
    "Сгибания рук с гантелями на скамье Скотта":               "Dumbbell_Preacher_Curl",
    "Молотки":                                                   "Hammer_Curls",
    "Сгибания рук с гантелями сидя":                            "Seated_Dumbbell_Curl",
    "Жим в хамере на верх груди":                               "Incline_Push-Up",
    "Жим в хамере на середину груди":                           "Dumbbell_Bench_Press",
    "Жим в хамере на низ груди":                                "Decline_Push-Up",
    "Жим штанги на горизонтальной скамье":                      "Barbell_Bench_Press_-_Medium_Grip",
    "Жим штанги на наклонной скамье":                           "Incline_Barbell_Bench_Press",
    "Жим штанги на скамье с обратным наклоном":                 "Decline_Barbell_Bench_Press",
    "Жим гантелей на наклонной скамье":                         "Incline_Dumbbell_Press",
    "Жим гантелей на горизонтальной скамье":                    "Dumbbell_Bench_Press",
    "Жим гантелей на скамье с обратным наклоном":               "Decline_Dumbbell_Press",
    "Отжимания на брусьях":                                      "Dips_-_Chest_Version",
    "Отжимания на брусьях с отягощением":                        "Dips_-_Chest_Version",
    "Сведение рук в кроссовере на верх груди":                   "Cable_Crossovers",
    "Сведение рук в кроссовере на середину груди":               "Cable_Crossovers",
    "Сведение рук в кроссовере на низ груди":                    "Cable_Crossovers",
    "Бабочка на грудь":                                          "Pec_Deck_Fly",
    "Подтягивания":                                              "Wide-Grip_Lat_Pulldown",
    "Подтягивания с отягощением":                                "Pull-Ups",
    "Горизонтальная тяга с узкой ручкой":                        "Seated_Cable_Rows",
    "Горизонтальная тяга с широкой ручкой":                      "Seated_Cable_Rows",
    "Вертикальная тяга с узкой ручкой":                          "Close-Grip_Front_Lat_Pulldown",
    "Вертикальная тяга с широкой ручкой":                        "Wide-Grip_Lat_Pulldown",
    "Т-гриф":                                                    "T-Bar_Row_with_Handle",
    "Тяга штанги к поясу":                                       "Bent_Over_Barbell_Row",
    "Тяга гантели к поясу":                                      "One-Arm_Dumbbell_Row",
    "Пуловер":                                                   "Dumbbell_Pullover",
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

HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
                  "(KHTML, like Gecko) Chrome/124.0 Safari/537.36",
    "Accept": "image/gif, image/jpeg, image/png, image/*, */*",
}


def get(url: str, timeout: int = 25) -> bytes | None:
    try:
        req = urllib.request.Request(url, headers=HEADERS)
        with urllib.request.urlopen(req, timeout=timeout) as r:
            data = r.read()
        return data if len(data) > 500 else None
    except Exception as e:
        print(f"      GET {url} FAIL: {type(e).__name__}: {e}")
        return None


def save(data: bytes, path: str) -> bool:
    try:
        with open(path, "wb") as f:
            f.write(data)
        return True
    except Exception as e:
        print(f"      WRITE FAIL: {e}")
        return False


def try_download(url: str, dest: str) -> bool:
    if os.path.isfile(dest) and os.path.getsize(dest) > 500:
        return True
    data = get(url)
    if data:
        save(data, dest)
        print(f"      OK  {os.path.basename(dest)}  ({len(data)//1024}KB)")
        return True
    return False





def load_free_exercise_db() -> dict:
    """Return dict: lowercase_id → exercise object"""
    cache_path = os.path.join(STATIC_DIR, "_exercises_db.json")
    if os.path.isfile(cache_path):
        with open(cache_path, encoding="utf-8") as f:
            raw = json.load(f)
    else:
        print("Downloading free-exercise-db exercises.json ...")
        data = get(EXERCISES_JSON_URL, timeout=60)
        if not data:
            print("FATAL: could not download exercises.json")
            return {}
        with open(cache_path, "wb") as f:
            f.write(data)
        raw = json.loads(data)
        print(f"  Downloaded {len(raw)} exercises from free-exercise-db")


    return {ex["id"].lower(): ex for ex in raw}







def load_wger_gif_map() -> dict[int, str]:
    """Returns dict: wger exercise_base_id → gif_url"""
    cache_path = os.path.join(STATIC_DIR, "_wger_images.json")
    if os.path.isfile(cache_path):
        with open(cache_path, encoding="utf-8") as f:
            all_images = json.load(f)
    else:
        print("Fetching wger exercise image list (paginated)...")
        all_images = []
        url = "https://wger.de/api/v2/exerciseimage/?format=json&limit=100&offset=0"
        while url:
            data = get(url, timeout=30)
            if not data:
                break
            page = json.loads(data)
            all_images.extend(page.get("results", []))
            url = page.get("next")
            time.sleep(0.2)
        with open(cache_path, "w", encoding="utf-8") as f:
            json.dump(all_images, f)
        print(f"  Fetched {len(all_images)} wger images total")

    gif_map: dict[int, str] = {}
    for img in all_images:
        img_url = img.get("image", "")
        if img_url.lower().endswith(".gif"):
            eid = img.get("exercise_base")
            if eid and eid not in gif_map:
                gif_map[eid] = img_url
    return gif_map


def wger_search_id(name: str) -> int | None:
    """Search wger for exercise by English name, return exercise_base id."""
    url = f"https://wger.de/api/v2/exercise/search/?term={urllib.parse.quote(name)}&format=json"
    data = get(url, timeout=15)
    if not data:
        return None
    try:
        j = json.loads(data)
        suggestions = j.get("suggestions", [])
        if suggestions:
            return suggestions[0].get("data", {}).get("base_id")
    except Exception:
        pass
    return None





KEYWORDS = {
    "жим": "press",
    "тяга": "row pull deadlift",
    "приседания": "squat",
    "сгибания": "curl",
    "разгибания": "extension",
    "ноги": "leg",
    "гантели": "dumbbell",
    "гантелями": "dumbbell",
    "штанги": "barbell",
    "штангой": "barbell",
    "тренажёре": "machine seated",
    "подтягивания": "pullup chin",
    "отжимания": "dip pushup",
    "горизонтальной": "flat bench",
    "наклонной": "incline",
    "скамье": "bench",
    "кроссовере": "cable",
    "плечи": "shoulder press",
    "грудь": "chest",
    "грудные": "chest",
    "спина": "back row",
    "ягодицы": "glute hip",
    "ягодичный": "glute hip thrust",
    "бицепс": "bicep curl",
    "трицепс": "tricep extension",
    "икры": "calf raise",
    "пресс": "ab crunch",
    "румынская": "romanian deadlift",
    "болгарские": "bulgarian split lunge",
    "выпады": "lunge split",
    "гиперэкстензия": "hyperextension back extension",
    "молотки": "hammer curl",
    "пуловер": "pullover",
    "скотта": "preacher",
    "бабочка": "pec deck fly",
    "французский": "skullcrusher tricep",
    "разведения": "abduction lateral",
    "сведения": "adduction cable",
    "махи": "lateral raise fly dumbbell",
    "отведение": "abduction kickback cable",
    "подъёмы": "calf raise",
    "носки": "calf raise toes",
    "вертикальная": "lat pulldown",
    "горизонтальная": "row seated cable",
    "широкой": "wide grip",
    "узкой": "close narrow grip",
    "сидя": "seated",
    "стоя": "standing",
    "лёжа": "lying",
    "т-гриф": "t-bar row",
}


def fuzzy_score(ru_name: str, ex: dict) -> int:
    en = (ex.get("name", "") + " " + " ".join(ex.get("primaryMuscles", []))).lower()
    score = 0
    for word in ru_name.lower().split():
        if word in KEYWORDS:
            for kw in KEYWORDS[word].split():
                if kw in en:
                    score += 2

    for word in ru_name.lower().split():
        if word in en:
            score += 1
    return score


def find_best_fuzzy(ru_name: str, db: dict) -> dict | None:
    best, best_score = None, 0
    for ex in db.values():
        s = fuzzy_score(ru_name, ex)
        if s > best_score:
            best_score, best = s, ex
    return best if best_score >= 2 else None





def main():
    print(f"Static dir: {STATIC_DIR}\n")

    db = load_free_exercise_db()
    if not db:
        print("ERROR: Empty exercise DB, aborting")
        return

    print("\nFetching wger GIF map...")
    gif_map = load_wger_gif_map()
    print(f"  {len(gif_map)} exercises with GIFs in wger\n")

    sql_lines = [
        "-- Auto-generated UPDATE statements for REAL exercise media",
        "-- Apply: psql -U sportapp_user -d sportapp_db -f backend/static/exercises/update_media_urls.sql\n",
    ]

    for our_id, ru_name in EXERCISES:
        print(f"[{our_id:02d}] {ru_name}")
        img_dest = os.path.join(STATIC_DIR, f"img_{our_id}.jpg")
        gif_dest = os.path.join(STATIC_DIR, f"vid_{our_id}.gif")


        db_ex = None
        manual_id = MANUAL_MAP.get(ru_name)
        if manual_id:
            db_ex = db.get(manual_id.lower())
            if not db_ex:

                db_ex = db.get(manual_id.replace("-", "_").lower())
            if not db_ex:

                for k, v in db.items():
                    if manual_id.lower() in k or k in manual_id.lower():
                        db_ex = v
                        break
        if not db_ex:
            db_ex = find_best_fuzzy(ru_name, db)
            if db_ex:
                print(f"      [fuzzy match] {db_ex['name']}")


        img_ok = False
        if db_ex:
            images = db_ex.get("images", [])
            for img_path in images:
                raw_url = f"{RAW_GH}/{img_path}"
                if try_download(raw_url, img_dest):
                    img_ok = True
                    break
        if not img_ok and os.path.isfile(img_dest) and os.path.getsize(img_dest) > 500:
            img_ok = True


        gif_ok = os.path.isfile(gif_dest) and os.path.getsize(gif_dest) > 500

        if not gif_ok:

            en_name = db_ex["name"] if db_ex else ru_name
            wger_id = wger_search_id(en_name)
            time.sleep(0.25)
            if wger_id and wger_id in gif_map:
                gif_ok = try_download(gif_map[wger_id], gif_dest)
            if not gif_ok:

                if wger_id:
                    info_url = f"https://wger.de/api/v2/exerciseinfo/{wger_id}/?format=json"
                    data = get(info_url, timeout=15)
                    if data:
                        try:
                            info = json.loads(data)
                            for img in info.get("images", []):
                                img_url = img.get("image", "")
                                if img_url:
                                    gif_ok = try_download(img_url, gif_dest)
                                    if gif_ok:
                                        break
                        except Exception:
                            pass
                time.sleep(0.2)


            if not gif_ok and img_ok:
                import shutil
                shutil.copy2(img_dest, gif_dest)
                print(f"      fallback: copied img → vid_{our_id}.gif")
                gif_ok = True

        image_url = f"/static/exercises/img_{our_id}.jpg" if img_ok else None
        video_url  = f"/static/exercises/vid_{our_id}.gif" if gif_ok else None

        if image_url and video_url:
            sql_lines.append(
                f"UPDATE sportapp.exercises "
                f"SET image_url = '{image_url}', video_url = '{video_url}' "
                f"WHERE id = {our_id};"
            )
        elif image_url:
            sql_lines.append(
                f"UPDATE sportapp.exercises SET image_url = '{image_url}' WHERE id = {our_id};"
            )

    sql_path = os.path.join(STATIC_DIR, "update_media_urls.sql")
    with open(sql_path, "w", encoding="utf-8") as f:
        f.write("\n".join(sql_lines) + "\n")

    imgs = len([f for f in os.listdir(STATIC_DIR) if f.startswith("img_") and f.endswith(".jpg")])
    gifs = len([f for f in os.listdir(STATIC_DIR) if f.startswith("vid_") and f.endswith(".gif")])
    print(f"\nDone!")
    print(f"  Real JPG images: {imgs}")
    print(f"  GIF/video files: {gifs}")
    print(f"  SQL file:        {sql_path}")


if __name__ == "__main__":
    main()