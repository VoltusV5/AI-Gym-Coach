
"""
fetch_exercise_media_v2.py
--------------------------
Strategy:
  1. For each exercise, search wger.de API by English name to get the real exercise ID.
  2. Download the animated GIF demo from wger (if available).
  3. Generate a beautiful branded SVG thumbnail locally (zero external deps).
  4. Print the SQL UPDATE statements to fill image_url / video_url in the DB.

Usage (from fitness-app root):
  python -X utf8 scripts/fetch_exercise_media_v2.py
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




EXERCISES = [
    (1,  "Leg Press"),
    (2,  "Barbell Squat"),
    (3,  "Hack Squat"),
    (4,  "Leg Extension"),
    (5,  "Bulgarian Split Squat"),
    (6,  "Leg Curl Seated"),
    (7,  "Leg Curl Lying"),
    (8,  "Romanian Deadlift"),
    (9,  "Romanian Deadlift Dumbbell"),
    (10, "Hyperextension"),
    (11, "Hip Thrust"),
    (12, "Hip Thrust Barbell"),
    (13, "Hip Abduction Machine"),
    (14, "Hip Adduction Machine"),
    (15, "Cable Kickback"),
    (16, "Glute Kickback"),
    (17, "Standing Calf Raise"),
    (18, "Seated Calf Raise"),
    (19, "Barbell Calf Raise"),
    (20, "Single Leg Calf Raise"),
    (21, "Calf Press"),
    (22, "Cable Crunch"),
    (23, "Kneeling Cable Crunch"),
    (24, "Dumbbell Lateral Raise"),
    (25, "Cable Lateral Raise"),
    (26, "Dumbbell Shoulder Press"),
    (27, "Shoulder Press Machine"),
    (28, "Reverse Fly Seated"),
    (29, "Reverse Fly Standing"),
    (30, "Rear Delt Fly Machine"),
    (31, "Triceps Pushdown"),
    (32, "Skull Crusher Dumbbell"),
    (33, "Skull Crusher Barbell"),
    (34, "Overhead Triceps Extension Dumbbell"),
    (35, "Overhead Triceps Extension Cable"),
    (36, "Cable Curl"),
    (37, "Barbell Curl"),
    (38, "Dumbbell Curl"),
    (39, "Preacher Curl Barbell"),
    (40, "Preacher Curl Dumbbell"),
    (41, "Hammer Curl"),
    (42, "Seated Dumbbell Curl"),
    (43, "Incline Chest Press Machine"),
    (44, "Chest Press Machine"),
    (45, "Decline Bench Press"),
    (46, "Flat Barbell Bench Press"),
    (47, "Incline Barbell Bench Press"),
    (48, "Decline Barbell Bench Press"),
    (49, "Incline Dumbbell Press"),
    (50, "Flat Dumbbell Press"),
    (51, "Decline Dumbbell Press"),
    (52, "Dips"),
    (53, "Weighted Dips"),
    (54, "Cable Crossover High"),
    (55, "Cable Crossover"),
    (56, "Cable Crossover Low"),
    (57, "Pec Deck Fly"),
    (58, "Pull-up"),
    (59, "Weighted Pull-up"),
    (60, "Seated Cable Row"),
    (61, "Seated Cable Row Wide"),
    (62, "Lat Pulldown Close Grip"),
    (63, "Lat Pulldown"),
    (64, "T-Bar Row"),
    (65, "Barbell Row"),
    (66, "Dumbbell Row"),
    (67, "Pullover"),
]

MUSCLE_GROUP_BY_ID = {
    **{i: "Ноги"   for i in range(1, 22)},
    **{i: "Пресс"  for i in [22, 23]},
    **{i: "Плечи"  for i in range(24, 31)},
    **{i: "Руки"   for i in range(31, 43)},
    **{i: "Грудь"  for i in range(43, 58)},
    **{i: "Спина"  for i in range(58, 68)},
}


PALETTES = {
    "Ноги":   ("
    "Пресс":  ("
    "Плечи":  ("
    "Руки":   ("
    "Грудь":  ("
    "Спина":  ("
}

ICONS = {
    "Ноги":   "🦵",
    "Пресс":  "💪",
    "Плечи":  "🏋️",
    "Руки":   "💪",
    "Грудь":  "🏋️",
    "Спина":  "🔙",
}

STATIC_DIR = os.path.normpath(
    os.path.join(os.path.dirname(__file__), "..", "backend", "static", "exercises")
)
WGER_BASE = "https://wger.de/api/v2"


def ensure_dir():
    os.makedirs(STATIC_DIR, exist_ok=True)


def download_file(url: str, dest: str, timeout: int = 20) -> bool:
    try:
        req = urllib.request.Request(url, headers={
            "User-Agent": "Mozilla/5.0 Sportik-App/1.0",
            "Accept": "*/*",
        })
        with urllib.request.urlopen(req, timeout=timeout) as resp, open(dest, "wb") as f:
            f.write(resp.read())
        size = os.path.getsize(dest)
        if size < 500:
            os.remove(dest)
            return False
        print(f"  OK  {os.path.basename(dest)} ({size // 1024}KB)")
        return True
    except Exception as e:
        print(f"  FAIL  {os.path.basename(dest)}: {type(e).__name__}: {e}")
        return False


def wger_search(name: str) -> int | None:
    """Search wger for an exercise by English name and return its ID."""
    url = f"{WGER_BASE}/exercise/search/?term={urllib.parse.quote(name)}&language=english&format=json"
    try:
        req = urllib.request.Request(url, headers={"User-Agent": "Sportik-App/1.0"})
        with urllib.request.urlopen(req, timeout=15) as resp:
            data = json.loads(resp.read())
        suggestions = data.get("suggestions", [])
        if suggestions:
            return suggestions[0].get("data", {}).get("id")
        return None
    except Exception as e:
        print(f"  search error: {e}")
        return None


def wger_exercise_gif(exercise_id: int) -> str | None:
    """Get the first GIF URL for a wger exercise ID."""
    url = f"{WGER_BASE}/exerciseinfo/{exercise_id}/?format=json"
    try:
        req = urllib.request.Request(url, headers={"User-Agent": "Sportik-App/1.0"})
        with urllib.request.urlopen(req, timeout=15) as resp:
            info = json.loads(resp.read())
        images = info.get("images", []) or []
        for img in images:
            img_url = img.get("image", "")
            if img_url:
                return img_url
        return None
    except Exception as e:
        print(f"  exerciseinfo error: {e}")
        return None


def generate_svg_image(ex_id: int, name: str, muscle_group: str) -> str:
    """Generate a beautiful branded SVG thumbnail for an exercise."""
    palette = PALETTES.get(muscle_group, ("
    c1, c2, accent = palette

    words = name.split()
    line1 = " ".join(words[:3])
    line2 = " ".join(words[3:]) if len(words) > 3 else ""

    svg = f"""<svg xmlns="http://www.w3.org/2000/svg" width="400" height="300" viewBox="0 0 400 300">
  <defs>
    <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%" stop-color="{c1}"/>
      <stop offset="100%" stop-color="{c2}"/>
    </linearGradient>
    <linearGradient id="bar" x1="0" y1="0" x2="1" y2="0">
      <stop offset="0%" stop-color="{accent}" stop-opacity="0"/>
      <stop offset="50%" stop-color="{accent}"/>
      <stop offset="100%" stop-color="{accent}" stop-opacity="0"/>
    </linearGradient>
    <filter id="glow">
      <feGaussianBlur stdDeviation="6" result="blur"/>
      <feMerge><feMergeNode in="blur"/><feMergeNode in="SourceGraphic"/></feMerge>
    </filter>
  </defs>
  <!-- Background -->
  <rect width="400" height="300" fill="url(
  <!-- Decorative circles -->
  <circle cx="360" cy="40" r="80" fill="{accent}" fill-opacity="0.06"/>
  <circle cx="40" cy="260" r="60" fill="{accent}" fill-opacity="0.05"/>
  <!-- Accent bar -->
  <rect x="80" y="140" width="240" height="2" fill="url(
  <!-- Exercise ID badge -->
  <rect x="16" y="16" width="44" height="24" rx="12" fill="{accent}" fill-opacity="0.2"/>
  <text x="38" y="32" font-family="system-ui,sans-serif" font-size="11" font-weight="700"
        fill="{accent}" text-anchor="middle">
  <!-- Muscle group label -->
  <text x="384" y="30" font-family="system-ui,sans-serif" font-size="11" font-weight="600"
        fill="{accent}" text-anchor="end" opacity="0.8">{muscle_group.upper()}</text>
  <!-- Exercise name line 1 -->
  <text x="200" y="{"130" if line2 else "155"}" font-family="system-ui,sans-serif" font-size="22" font-weight="700"
        fill="
  {"" if not line2 else f'<text x="200" y="162" font-family="system-ui,sans-serif" font-size="22" font-weight="700" fill="
  <!-- Accent dot row -->
  <circle cx="190" cy="185" r="3" fill="{accent}" opacity="0.5"/>
  <circle cx="200" cy="185" r="3" fill="{accent}" opacity="0.8"/>
  <circle cx="210" cy="185" r="3" fill="{accent}" opacity="0.5"/>
  <!-- Brand watermark -->
  <text x="200" y="275" font-family="system-ui,sans-serif" font-size="10" fill="{accent}"
        text-anchor="middle" opacity="0.35">SPORTIK • AI FITNESS</text>
</svg>"""
    return svg


def generate_svg_animation(ex_id: int, name: str, muscle_group: str) -> str:
    """Generate a simple animated SVG that simulates a GIF animation."""
    palette = PALETTES.get(muscle_group, ("
    c1, c2, accent = palette

    svg = f"""<svg xmlns="http://www.w3.org/2000/svg" width="400" height="300" viewBox="0 0 400 300">
  <defs>
    <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%" stop-color="{c1}"/>
      <stop offset="100%" stop-color="{c2}"/>
    </linearGradient>
    <style>
      @keyframes pulse {{
        0%,100% {{ opacity:1; transform:scale(1); }}
        50% {{ opacity:0.6; transform:scale(0.93); }}
      }}
      @keyframes slideUp {{
        0%,100% {{ transform:translateY(0px); }}
        50% {{ transform:translateY(-12px); }}
      }}
      @keyframes spin {{
        from {{ transform:rotate(0deg); }}
        to {{ transform:rotate(360deg); }}
      }}
      @keyframes blink {{
        0%,100% {{ opacity:1; }}
        50% {{ opacity:0.3; }}
      }}
      .figure {{ animation:slideUp 2s ease-in-out infinite; transform-origin:200px 160px; }}
      .glow {{ animation:pulse 2s ease-in-out infinite; }}
      .ring {{ animation:spin 4s linear infinite; transform-origin:200px 105px; }}
      .dots {{ animation:blink 1.2s ease-in-out infinite; }}
    </style>
  </defs>
  <!-- Background -->
  <rect width="400" height="300" fill="url(
  <!-- Soft glow circle -->
  <circle cx="200" cy="105" r="65" fill="{accent}" fill-opacity="0.12" class="glow"/>
  <!-- Spinning ring -->
  <circle cx="200" cy="105" r="58" fill="none" stroke="{accent}" stroke-width="2"
          stroke-dasharray="30 10" opacity="0.4" class="ring"/>
  <!-- Stick figure body (animated up/down) -->
  <g class="figure">
    <!-- head -->
    <circle cx="200" cy="70" r="16" fill="none" stroke="{accent}" stroke-width="2.5"/>
    <!-- body -->
    <line x1="200" y1="86" x2="200" y2="130" stroke="{accent}" stroke-width="2.5" stroke-linecap="round"/>
    <!-- arms raised -->
    <line x1="200" y1="100" x2="172" y2="118" stroke="{accent}" stroke-width="2.5" stroke-linecap="round"/>
    <line x1="200" y1="100" x2="228" y2="118" stroke="{accent}" stroke-width="2.5" stroke-linecap="round"/>
    <!-- legs -->
    <line x1="200" y1="130" x2="182" y2="158" stroke="{accent}" stroke-width="2.5" stroke-linecap="round"/>
    <line x1="200" y1="130" x2="218" y2="158" stroke="{accent}" stroke-width="2.5" stroke-linecap="round"/>
  </g>
  <!-- Exercise name -->
  <text x="200" y="200" font-family="system-ui,sans-serif" font-size="16" font-weight="700"
        fill="
  <text x="200" y="220" font-family="system-ui,sans-serif" font-size="11" font-weight="600"
        fill="{accent}" text-anchor="middle" opacity="0.8">{muscle_group.upper()}</text>
  <!-- Animated dots -->
  <g class="dots">
    <circle cx="187" cy="240" r="3" fill="{accent}"/>
    <circle cx="200" cy="240" r="3" fill="{accent}"/>
    <circle cx="213" cy="240" r="3" fill="{accent}"/>
  </g>
  <!-- Brand -->
  <text x="200" y="280" font-family="system-ui,sans-serif" font-size="9" fill="{accent}"
        text-anchor="middle" opacity="0.3">SPORTIK • AI FITNESS</text>
</svg>"""
    return svg


def main():
    ensure_dir()
    print(f"Static dir: {STATIC_DIR}\n")

    sql_lines = [
        "-- Auto-generated UPDATE statements for exercise media",
        "-- Apply in psql or your migration tool:\n",
    ]

    for our_id, name in EXERCISES:
        muscle_group = MUSCLE_GROUP_BY_ID.get(our_id, "Ноги")
        img_path = os.path.join(STATIC_DIR, f"img_{our_id}.svg")
        gif_path = os.path.join(STATIC_DIR, f"vid_{our_id}.gif")
        gif_svg_path = os.path.join(STATIC_DIR, f"vid_{our_id}.svg")

        print(f"[{our_id:02d}] {name}")


        if not os.path.isfile(img_path):
            svg_content = generate_svg_image(our_id, name, muscle_group)
            with open(img_path, "w", encoding="utf-8") as f:
                f.write(svg_content)
            print(f"  GEN img_{our_id}.svg")


        gif_exists = os.path.isfile(gif_path)
        gif_svg_exists = os.path.isfile(gif_svg_path)

        if not gif_exists and not gif_svg_exists:

            wger_id = wger_search(name)
            time.sleep(0.3)
            gif_downloaded = False
            if wger_id:
                gif_url = wger_exercise_gif(wger_id)
                if gif_url:
                    gif_downloaded = download_file(gif_url, gif_path)
                time.sleep(0.2)

            if not gif_downloaded:

                svg_anim = generate_svg_animation(our_id, name, muscle_group)
                with open(gif_svg_path, "w", encoding="utf-8") as f:
                    f.write(svg_anim)
                print(f"  GEN vid_{our_id}.svg (animated)")
        else:
            status = "GIF cached" if gif_exists else "SVG cached"
            print(f"  already {status}")


        video_url = (
            f"/static/exercises/vid_{our_id}.gif"
            if os.path.isfile(gif_path)
            else f"/static/exercises/vid_{our_id}.svg"
        )
        image_url = f"/static/exercises/img_{our_id}.svg"

        sql_lines.append(
            f"UPDATE sportapp.exercises "
            f"SET image_url = '{image_url}', "
            f"    video_url = '{video_url}' "
            f"WHERE id = {our_id};"
        )

    sql_path = os.path.join(STATIC_DIR, "update_media_urls.sql")
    with open(sql_path, "w", encoding="utf-8") as f:
        f.write("\n".join(sql_lines) + "\n")

    print(f"\nDone! SQL written to: {sql_path}")

    gifs = len([f for f in os.listdir(STATIC_DIR) if f.endswith(".gif")])
    svgs = len([f for f in os.listdir(STATIC_DIR) if f.endswith(".svg")])
    print(f"  Real GIFs downloaded: {gifs}")
    print(f"  SVG files generated:  {svgs}")


if __name__ == "__main__":
    main()