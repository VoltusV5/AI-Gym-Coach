
"""
fetch_exercise_media_v3.py
--------------------------
Downloads real exercise JPG thumbnails and animated GIFs from
  https://github.com/hasaneyldrm/exercises-dataset
using fuzzy name matching to pair our 67 exercises with the dataset.

Usage (from fitness-app root):
  python -X utf8 scripts/fetch_exercise_media_v3.py

Output:
  backend/static/exercises/img_<id>.jpg   — thumbnail photo
  backend/static/exercises/vid_<id>.gif   — animated demonstration
  backend/static/exercises/update_media_urls.sql
"""

import os
import sys
import io
import json
import time
import urllib.request
import urllib.error
from difflib import SequenceMatcher


if sys.stdout.encoding != "utf-8":
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8", errors="replace")




OUR_EXERCISES = [
    (1,  "leg press"),
    (2,  "barbell squat"),
    (3,  "hack squat"),
    (4,  "leg extension"),
    (5,  "bulgarian split squat"),
    (6,  "seated leg curl"),
    (7,  "lying leg curl"),
    (8,  "romanian deadlift barbell"),
    (9,  "romanian deadlift dumbbell"),
    (10, "hyperextension"),
    (11, "hip thrust dumbbell"),
    (12, "hip thrust barbell"),
    (13, "hip abduction machine"),
    (14, "hip adduction machine"),
    (15, "cable kickback"),
    (16, "glute kickback"),
    (17, "standing calf raise"),
    (18, "seated calf raise"),
    (19, "barbell calf raise"),
    (20, "single leg calf raise"),
    (21, "calf press"),
    (22, "cable crunch"),
    (23, "kneeling cable crunch"),
    (24, "dumbbell lateral raise"),
    (25, "cable lateral raise"),
    (26, "dumbbell shoulder press"),
    (27, "shoulder press machine"),
    (28, "reverse fly seated dumbbell"),
    (29, "reverse fly standing"),
    (30, "rear delt fly machine"),
    (31, "triceps pushdown cable"),
    (32, "skull crusher dumbbell"),
    (33, "skull crusher barbell"),
    (34, "overhead triceps extension dumbbell"),
    (35, "overhead triceps extension cable"),
    (36, "cable curl biceps"),
    (37, "barbell curl biceps"),
    (38, "dumbbell curl biceps"),
    (39, "preacher curl barbell"),
    (40, "preacher curl dumbbell"),
    (41, "hammer curl"),
    (42, "seated dumbbell curl"),
    (43, "incline chest press machine"),
    (44, "chest press machine"),
    (45, "decline bench press dumbbell"),
    (46, "barbell bench press flat"),
    (47, "incline barbell bench press"),
    (48, "decline barbell bench press"),
    (49, "incline dumbbell press chest"),
    (50, "flat dumbbell press chest"),
    (51, "decline dumbbell press chest"),
    (52, "dips chest"),
    (53, "weighted dips"),
    (54, "cable crossover high pulley"),
    (55, "cable crossover chest"),
    (56, "cable crossover low pulley"),
    (57, "pec deck fly machine"),
    (58, "pull up"),
    (59, "weighted pull up"),
    (60, "seated cable row"),
    (61, "seated cable row wide grip"),
    (62, "lat pulldown close grip"),
    (63, "lat pulldown wide grip"),
    (64, "t bar row"),
    (65, "barbell row"),
    (66, "dumbbell row"),
    (67, "pullover dumbbell"),
]





MANUAL_HINTS = {
    1:  "leg press",
    2:  "barbell squat",
    3:  "hack squat",
    4:  "leg extension",
    5:  "bulgarian split squat",
    6:  "seated leg curl",
    7:  "lying leg curl",
    8:  "romanian deadlift",
    9:  "romanian deadlift dumbbell",
    10: "hyperextension",
    11: "hip thrust",
    12: "barbell hip thrust",
    13: "hip abduction",
    14: "hip adduction",
    15: "cable kickback",
    16: "glute kickback",
    17: "standing calf raise",
    18: "seated calf raise",
    19: "calf raise barbell",
    20: "single leg calf raise",
    21: "calf press",
    22: "cable crunch",
    23: "kneeling cable crunch",
    24: "dumbbell lateral raise",
    25: "cable lateral raise",
    26: "dumbbell shoulder press",
    27: "shoulder press machine",
    28: "seated dumbbell reverse fly",
    29: "standing reverse fly",
    30: "rear delt machine",
    31: "triceps pushdown",
    32: "skull crusher",
    33: "barbell skull crusher",
    34: "overhead triceps extension",
    35: "cable overhead triceps extension",
    36: "cable curl",
    37: "barbell curl",
    38: "dumbbell curl",
    39: "preacher curl",
    40: "dumbbell preacher curl",
    41: "hammer curl",
    42: "seated dumbbell curl",
    43: "incline chest press machine",
    44: "chest press machine",
    45: "decline bench press",
    46: "barbell bench press",
    47: "incline barbell bench press",
    48: "decline barbell bench press",
    49: "incline dumbbell press",
    50: "dumbbell bench press",
    51: "decline dumbbell press",
    52: "dips",
    53: "weighted dips",
    54: "cable crossover",
    55: "cable fly",
    56: "cable crossover low",
    57: "pec deck",
    58: "pull-up",
    59: "weighted pull-up",
    60: "seated cable row",
    61: "cable row wide",
    62: "lat pulldown close grip",
    63: "lat pulldown",
    64: "t-bar row",
    65: "barbell row",
    66: "dumbbell row",
    67: "dumbbell pullover",
}


GITHUB_BASE = "https://raw.githubusercontent.com/hasaneyldrm/exercises-dataset/main"
DATASET_JSON_URL = f"{GITHUB_BASE}/data/exercises.json"

STATIC_DIR = os.path.normpath(
    os.path.join(os.path.dirname(__file__), "..", "backend", "static", "exercises")
)


def ensure_dir():
    os.makedirs(STATIC_DIR, exist_ok=True)


def download_json(url: str) -> list:
    print(f"Fetching dataset from:\n  {url}")
    req = urllib.request.Request(url, headers={
        "User-Agent": "Mozilla/5.0 Sportik-App/1.0",
        "Accept": "application/json",
    })
    with urllib.request.urlopen(req, timeout=30) as resp:
        raw = resp.read().decode("utf-8")
    data = json.loads(raw)
    print(f"  Loaded {len(data)} exercises from dataset.\n")
    return data


def similarity(a: str, b: str) -> float:
    return SequenceMatcher(None, a.lower(), b.lower()).ratio()


def find_best_match(hint: str, dataset: list) -> dict | None:
    """Return the dataset entry with the highest name similarity to hint."""
    best_score = 0.0
    best_entry = None
    hint_lower = hint.lower()
    for entry in dataset:
        name_lower = entry["name"].lower()
        score = similarity(hint_lower, name_lower)

        if hint_lower in name_lower or name_lower in hint_lower:
            score = max(score, 0.75)

        hint_words = set(hint_lower.split())
        name_words = set(name_lower.split())
        common = hint_words & name_words
        if common:
            word_score = len(common) / max(len(hint_words), len(name_words))
            score = max(score, word_score * 0.9)
        if score > best_score:
            best_score = score
            best_entry = entry
    return best_entry, best_score


def download_file(url: str, dest: str, timeout: int = 30) -> bool:
    try:
        req = urllib.request.Request(url, headers={
            "User-Agent": "Mozilla/5.0 Sportik-App/1.0",
        })
        with urllib.request.urlopen(req, timeout=timeout) as resp, open(dest, "wb") as f:
            data = resp.read()
            f.write(data)
        size = os.path.getsize(dest)
        if size < 500:
            os.remove(dest)
            return False
        print(f"    ✓ {os.path.basename(dest)} ({size // 1024} KB)")
        return True
    except Exception as e:
        print(f"    ✗ {os.path.basename(dest)}: {type(e).__name__}: {e}")
        if os.path.exists(dest):
            os.remove(dest)
        return False


def main():
    ensure_dir()
    print(f"Static dir: {STATIC_DIR}\n")


    try:
        dataset = download_json(DATASET_JSON_URL)
    except Exception as e:
        print(f"ERROR fetching dataset JSON: {e}")
        sys.exit(1)

    sql_lines = [
        "-- Auto-generated UPDATE statements for exercise media (v3)",
        "-- Source: github.com/hasaneyldrm/exercises-dataset",
        "-- Apply: psql -f update_media_urls.sql\n",
    ]

    matched = 0
    failed_img = 0
    failed_gif = 0

    for our_id, our_name in OUR_EXERCISES:
        hint = MANUAL_HINTS.get(our_id, our_name)
        best_entry, score = find_best_match(hint, dataset)

        print(f"[{our_id:02d}] {our_name}")
        if best_entry:
            print(f"      → Dataset match: \"{best_entry['name']}\" (score={score:.2f})")
        else:
            print(f"      → NO MATCH FOUND")
            continue

        img_dest = os.path.join(STATIC_DIR, f"img_{our_id}.jpg")
        gif_dest = os.path.join(STATIC_DIR, f"vid_{our_id}.gif")


        img_ok = False
        if os.path.isfile(img_dest) and os.path.getsize(img_dest) > 500:
            print(f"    ↩ img_{our_id}.jpg already cached")
            img_ok = True
        elif best_entry.get("image"):
            img_url = f"{GITHUB_BASE}/{best_entry['image']}"
            img_ok = download_file(img_url, img_dest)
            if img_ok:
                matched += 1
            else:
                failed_img += 1
            time.sleep(0.15)


        gif_ok = False
        if os.path.isfile(gif_dest) and os.path.getsize(gif_dest) > 500:
            print(f"    ↩ vid_{our_id}.gif already cached")
            gif_ok = True
        elif best_entry.get("gif_url"):
            gif_url = f"{GITHUB_BASE}/{best_entry['gif_url']}"
            gif_ok = download_file(gif_url, gif_dest)
            if not gif_ok:
                failed_gif += 1
            time.sleep(0.15)


        image_url = f"/static/exercises/img_{our_id}.jpg" if img_ok else None
        video_url = f"/static/exercises/vid_{our_id}.gif" if gif_ok else None

        if image_url and video_url:
            sql_lines.append(
                f"UPDATE sportapp.exercises "
                f"SET image_url = '{image_url}', "
                f"    video_url = '{video_url}' "
                f"WHERE id = {our_id};"
            )
        elif image_url:
            sql_lines.append(
                f"UPDATE sportapp.exercises "
                f"SET image_url = '{image_url}' "
                f"WHERE id = {our_id};"
            )
        print()


    sql_path = os.path.join(STATIC_DIR, "update_media_urls.sql")
    with open(sql_path, "w", encoding="utf-8") as f:
        f.write("\n".join(sql_lines) + "\n")


    gifs  = len([f for f in os.listdir(STATIC_DIR) if f.endswith(".gif")])
    imgs  = len([f for f in os.listdir(STATIC_DIR) if f.endswith(".jpg")])
    svgs  = len([f for f in os.listdir(STATIC_DIR) if f.endswith(".svg")])

    print("=" * 60)
    print(f"Done! SQL written to: {sql_path}")
    print(f"  JPG thumbnails : {imgs}")
    print(f"  Real GIFs      : {gifs}")
    print(f"  SVG fallbacks  : {svgs}")
    print(f"  Download fails : img={failed_img}, gif={failed_gif}")
    print("=" * 60)
    print("\nNext step — apply the SQL:")
    print('  psql -U postgres -d sportapp -f backend/static/exercises/update_media_urls.sql')


if __name__ == "__main__":
    main()