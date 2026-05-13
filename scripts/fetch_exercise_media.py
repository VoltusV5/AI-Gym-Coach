
"""
fetch_exercise_media.py
-----------------------
Downloads real exercise GIF animations and JPG thumbnails from the
wger open-source fitness database (https://wger.de/api/v2/) for every
exercise stored in the Sportik app.

Saves:
  backend/static/exercises/img_{id}.jpg  — thumbnail/image
  backend/static/exercises/vid_{id}.gif  — animated technique GIF

Then prints the SQL UPDATE statements to apply to the DB
so that image_url / video_url columns are filled.

Usage (from fitness-app root):
  python scripts/fetch_exercise_media.py
"""

import os
import time
import urllib.request
import json
import urllib.parse









EXERCISES = [

    (1,  72,  "Leg Press"),
    (2,  69,  "Barbell Squat"),
    (3,  70,  "Hack Squat"),
    (4,  73,  "Leg Extension"),
    (5,  74,  "Bulgarian Split Squat"),
    (6,  75,  "Leg Curl Seated"),
    (7,  76,  "Leg Curl Lying"),
    (8,  77,  "Romanian Deadlift Barbell"),
    (9,  78,  "Romanian Deadlift Dumbbell"),
    (10, 79,  "Hyperextension"),
    (11, 80,  "Hip Thrust Machine"),
    (12, 81,  "Hip Thrust Barbell"),
    (13, 82,  "Abduction Machine"),
    (14, 83,  "Adduction Machine"),
    (15, 84,  "Cable Kickback"),
    (16, 85,  "Glute Kickback Machine"),
    (17, 86,  "Standing Calf Raise Machine"),
    (18, 87,  "Seated Calf Raise Machine"),
    (19, 88,  "Barbell Calf Raise"),
    (20, 89,  "Single Leg Calf Raise"),
    (21, 90,  "Calf Press Leg Press Machine"),
    (22, 91,  "Cable Crunch"),
    (23, 92,  "Cable Pullover Crunch"),
    (24, 93,  "Dumbbell Lateral Raise"),
    (25, 94,  "Cable Lateral Raise"),
    (26, 95,  "Dumbbell Shoulder Press"),
    (27, 96,  "Shoulder Press Machine"),
    (28, 97,  "Reverse Fly Seated Dumbbell"),
    (29, 98,  "Reverse Fly Standing Dumbbell"),
    (30, 99,  "Reverse Fly Machine"),
    (31, 100, "Triceps Pushdown Cable"),
    (32, 101, "Skull Crusher Dumbbell"),
    (33, 102, "Skull Crusher Barbell"),
    (34, 103, "Overhead Triceps Extension Dumbbell"),
    (35, 104, "Overhead Triceps Extension Cable"),
    (36, 105, "Cable Curl"),
    (37, 106, "Barbell Curl"),
    (38, 107, "Dumbbell Curl Standing"),
    (39, 108, "Preacher Curl Barbell"),
    (40, 109, "Preacher Curl Dumbbell"),
    (41, 110, "Hammer Curl"),
    (42, 111, "Dumbbell Curl Seated"),
    (43, 112, "Incline Chest Press Machine"),
    (44, 113, "Chest Press Machine"),
    (45, 114, "Decline Chest Press Machine"),
    (46, 115, "Flat Barbell Bench Press"),
    (47, 116, "Incline Barbell Bench Press"),
    (48, 117, "Decline Barbell Bench Press"),
    (49, 118, "Incline Dumbbell Press"),
    (50, 119, "Flat Dumbbell Press"),
    (51, 120, "Decline Dumbbell Press"),
    (52, 121, "Dips"),
    (53, 122, "Weighted Dips"),
    (54, 123, "Cable Crossover High"),
    (55, 124, "Cable Crossover Mid"),
    (56, 125, "Cable Crossover Low"),
    (57, 126, "Pec Deck Fly"),
    (58, 127, "Pull-up"),
    (59, 128, "Weighted Pull-up"),
    (60, 129, "Seated Cable Row Close Grip"),
    (61, 130, "Seated Cable Row Wide Grip"),
    (62, 131, "Lat Pulldown Close Grip"),
    (63, 132, "Lat Pulldown Wide Grip"),
    (64, 133, "T-Bar Row"),
    (65, 134, "Barbell Row"),
    (66, 135, "Dumbbell Row"),
    (67, 136, "Pullover"),
]



MUSCLE_GROUP_KEYWORDS = {
    "Ноги":   "leg workout gym exercise",
    "Пресс":  "ab crunch gym exercise",
    "Плечи":  "shoulder dumbbell gym exercise",
    "Руки":   "bicep curl gym exercise",
    "Грудь":  "chest press bench gym",
    "Спина":  "back row pulldown gym",
}

STATIC_DIR = os.path.join(os.path.dirname(__file__), "..", "backend", "static", "exercises")
STATIC_DIR = os.path.normpath(STATIC_DIR)


WGER_BASE = "https://wger.de/api/v2"


import sys
import io


if sys.stdout.encoding != 'utf-8':
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8', errors='replace')


def ensure_dir():
    os.makedirs(STATIC_DIR, exist_ok=True)


def download_file(url: str, dest: str, timeout: int = 15) -> bool:
    """Download a file and return True on success."""
    try:
        req = urllib.request.Request(url, headers={"User-Agent": "Sportik-App/1.0"})
        with urllib.request.urlopen(req, timeout=timeout) as resp, open(dest, "wb") as f:
            f.write(resp.read())
        print(f"  OK  {os.path.basename(dest)}")
        return True
    except Exception as e:
        print(f"  FAIL  {os.path.basename(dest)}: {e}")
        return False


def get_wger_exercise_info(wger_id: int) -> dict | None:
    """Fetch exercise info JSON from wger API."""
    url = f"{WGER_BASE}/exerciseinfo/{wger_id}/?format=json"
    try:
        req = urllib.request.Request(url, headers={"User-Agent": "Sportik-App/1.0"})
        with urllib.request.urlopen(req, timeout=15) as resp:
            return json.loads(resp.read())
    except Exception as e:
        print(f"  wger API error for id={wger_id}: {e}")
        return None


def get_exercise_gif_url(info: dict) -> str | None:
    """Extract the first animated GIF url from wger exerciseinfo response."""
    images = info.get("images", []) or []
    for img in images:
        img_url = img.get("image", "")
        if img_url.lower().endswith(".gif"):
            return img_url
    for img in images:
        img_url = img.get("image", "")
        if img_url:
            return img_url
    return None


def get_exercise_thumbnail_url(info: dict) -> str | None:
    """Extract first JPEG/PNG thumbnail from wger exerciseinfo response."""
    images = info.get("images", []) or []
    for img in images:
        img_url = img.get("image", "")
        if img_url and not img_url.lower().endswith(".gif"):
            return img_url
    for img in images:
        img_url = img.get("image", "")
        if img_url:
            return img_url
    return None


def fetch_unsplash_placeholder(exercise_name: str, dest: str) -> bool:
    """Fetch a representative photo from picsum as placeholder."""
    seed = abs(hash(exercise_name)) % 1000
    url = f"https://picsum.photos/seed/{seed}/400/300"
    return download_file(url, dest, timeout=10)


def main():
    ensure_dir()
    print(f"Static dir: {STATIC_DIR}\n")

    sql_lines = [
        "-- Auto-generated UPDATE statements for exercise media",
        "-- Apply in psql or your migration tool:\n",
    ]

    for our_id, wger_id, name in EXERCISES:
        img_dest = os.path.join(STATIC_DIR, f"img_{our_id}.jpg")
        gif_dest = os.path.join(STATIC_DIR, f"vid_{our_id}.gif")

        img_exists = os.path.isfile(img_dest)
        gif_exists = os.path.isfile(gif_dest)

        print(f"[{our_id:02d}] {name}")

        if not img_exists or not gif_exists:
            info = get_wger_exercise_info(wger_id)
            time.sleep(0.4)

            if info:
                gif_url = get_exercise_gif_url(info)
                img_url = get_exercise_thumbnail_url(info)

                if not gif_exists:
                    if gif_url:
                        download_file(gif_url, gif_dest)
                    else:

                        fetch_unsplash_placeholder(name, gif_dest)

                if not img_exists:
                    if img_url and img_url != gif_url:
                        download_file(img_url, img_dest)
                    else:
                        fetch_unsplash_placeholder(name, img_dest)
            else:

                if not gif_exists:
                    fetch_unsplash_placeholder(name, gif_dest)
                if not img_exists:
                    fetch_unsplash_placeholder(name, img_dest)
        else:
            print(f"  ✓  already cached")

        sql_lines.append(
            f"UPDATE sportapp.exercises "
            f"SET image_url = '/static/exercises/img_{our_id}.jpg', "
            f"    video_url = '/static/exercises/vid_{our_id}.gif' "
            f"WHERE id = {our_id};"
        )

    sql_path = os.path.join(STATIC_DIR, "update_media_urls.sql")
    with open(sql_path, "w", encoding="utf-8") as f:
        f.write("\n".join(sql_lines))

    print(f"\n✅ Done! SQL file written to: {sql_path}")
    print("   Apply it with: psql -U <user> -d <db> -f backend/static/exercises/update_media_urls.sql")
    print("\n--- SQL Preview (first 5) ---")
    for line in sql_lines[3:8]:
        print(line)


if __name__ == "__main__":
    main()