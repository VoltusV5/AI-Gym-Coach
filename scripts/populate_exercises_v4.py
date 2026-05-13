import json
import os
import shutil


DATASET_DIR = os.path.abspath("../exercises-dataset-main")
PROJECT_ROOT = os.path.abspath(".")
STATIC_DIR = os.path.join(PROJECT_ROOT, "backend", "static", "exercises")
JSON_PATH = os.path.join(DATASET_DIR, "data", "exercises.json")
SQL_OUTPUT = os.path.join(STATIC_DIR, "populate_v4.sql")

def main():
    if not os.path.exists(JSON_PATH):
        print(f"Error: Dataset JSON not found at {JSON_PATH}")
        return

    if not os.path.exists(STATIC_DIR):
        os.makedirs(STATIC_DIR, exist_ok=True)

    with open(JSON_PATH, "r", encoding="utf-8") as f:
        data = json.load(f)

    print(f"Loaded {len(data)} exercises from dataset.")

    sql_statements = []


    count = min(67, len(data))

    for i in range(count):
        entry = data[i]
        our_id = i + 1


        img_rel_path = entry.get("image")
        if img_rel_path:
            img_src = os.path.join(DATASET_DIR, img_rel_path)
            img_dst_name = f"img_{our_id}.jpg"
            img_dst = os.path.join(STATIC_DIR, img_dst_name)

            if os.path.exists(img_src):
                shutil.copy2(img_src, img_dst)
                image_url = f"/static/exercises/{img_dst_name}"
            else:
                print(f"Warning: Image {img_src} not found.")
                image_url = None
        else:
            image_url = None


        vid_rel_path = entry.get("gif_url")
        if vid_rel_path:
            vid_src = os.path.join(DATASET_DIR, vid_rel_path)
            vid_dst_name = f"vid_{our_id}.gif"
            vid_dst = os.path.join(STATIC_DIR, vid_dst_name)

            if os.path.exists(vid_src):
                shutil.copy2(vid_src, vid_dst)
                video_url = f"/static/exercises/{vid_dst_name}"
            else:
                print(f"Warning: Video {vid_src} not found.")
                video_url = None
        else:
            video_url = None


        if image_url and video_url:
            sql = f"UPDATE sportapp.exercises SET image_url = '{image_url}', video_url = '{video_url}' WHERE id = {our_id};"
        elif image_url:
            sql = f"UPDATE sportapp.exercises SET image_url = '{image_url}' WHERE id = {our_id};"
        elif video_url:
            sql = f"UPDATE sportapp.exercises SET video_url = '{video_url}' WHERE id = {our_id};"
        else:
            continue

        sql_statements.append(sql)

    with open(SQL_OUTPUT, "w", encoding="utf-8") as f:
        f.write("\n".join(sql_statements))
        f.write("\n")

    print(f"Copied files and generated SQL for {len(sql_statements)} exercises.")
    print(f"SQL file saved to: {SQL_OUTPUT}")

if __name__ == "__main__":
    main()