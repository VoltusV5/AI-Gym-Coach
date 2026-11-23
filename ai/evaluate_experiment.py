import argparse
import json
import os
import random
import time
import numpy as np
import pandas as pd
import torch
from catboost import CatBoostClassifier
from sklearn.metrics import accuracy_score, balanced_accuracy_score, f1_score
from sklearn.model_selection import train_test_split
from transformers import AutoModelForCausalLM, AutoTokenizer

SEED = 42
random.seed(SEED)
np.random.seed(SEED)
torch.manual_seed(SEED)

DATASET_PATH = "dataset_2.jsonlines"
CATBOOST_MODEL_PATH = "workout_model.cbm"
BASE_TRANSFORMER_NAME = "Qwen/Qwen2.5-0.5B-Instruct"
FINETUNED_TRANSFORMER_DIR = "models/qwen2.5-0.5b-workout-ft"

SPLIT_MAP_RU_TO_ID = {
    "Фулбади": 0,
    "Фулбоди": 0,
    "full_body": 0,
    "ВерхНиз": 1,
    "upper_lower": 1,
    "Тяни-толкай": 2,
    "push_pull": 2,
    "Пуш-пул-ноги": 3,
    "пуш-пул-ноги": 3,
    "push_pull_legs": 3,
    "Сплит": 4,
    "split": 4,
}

EXPECTED_DAYS_BY_SPLIT_ID = {
    0: 3,
    1: 4,
    2: 3,
    3: 6,
    4: 4,
}


def extract_json_object(text: str) -> dict | None:
    """Находит первый сбалансированный {...} и парсит JSON."""
    start = text.find("{")
    if start == -1:
        return None
    depth = 0
    in_str = False
    esc = False
    quote = None
    for i in range(start, len(text)):
        ch = text[i]
        if in_str:
            if esc:
                esc = False
            elif ch == "\\":
                esc = True
            elif ch == quote:
                in_str = False
            continue
        if ch in "\"'":
            in_str = True
            quote = ch
            continue
        if ch == "{":
            depth += 1
        elif ch == "}":
            depth -= 1
            if depth == 0:
                try:
                    return json.loads(text[start : i + 1])
                except json.JSONDecodeError:
                    return None
    return None


def load_and_split_data(path: str):
    """Загрузка данных, очистка дублей и фиксированный split без утечек."""
    if not os.path.exists(path):
        raise FileNotFoundError(f"Не найден файл датасета: {path}")

    rows = []
    seen_inputs = set()
    with open(path, "r", encoding="utf-8") as f:
        for line in f:
            if not line.strip():
                continue
            obj = json.loads(line)
            if not isinstance(obj.get("input"), dict) or not isinstance(obj.get("output"), dict):
                continue
            sig = json.dumps(obj["input"], sort_keys=True, ensure_ascii=False)
            if sig not in seen_inputs:
                seen_inputs.add(sig)
                rows.append(obj)

    train_val, test_data = train_test_split(rows, test_size=0.15, random_state=SEED)
    train_data, val_data = train_test_split(train_val, test_size=0.15, random_state=SEED)
    return train_data, val_data, test_data


def evaluate_catboost_baseline(test_data: list) -> dict:
    """Оценка зафиксированного baseline (CatBoost + расчёт совпадения дней)."""
    if not os.path.exists(CATBOOST_MODEL_PATH):
        raise FileNotFoundError(f"Не найден файл модели CatBoost: {CATBOOST_MODEL_PATH}")

    model = CatBoostClassifier()
    model.load_model(CATBOOST_MODEL_PATH)

    records, y_true, days_matches = [], [], []
    for item in test_data:
        inp = item["input"]
        out = item["output"]
        y_true.append(SPLIT_MAP_RU_TO_ID.get(out.get("тип_сплита"), 0))
        req_days = len(inp.get("дни_тренировок", []))
        records.append(
            {
                "возраст": inp["возраст"],
                "рост": inp["рост"],
                "вес": inp["вес"],
                "пол_num": 0 if inp["пол"] == "м" else 1,
                "тип_активности": inp["тип_активности"],
                "травмы_num": 0 if inp["травмы_или_болезни"] == "нет" else 1,
                "цель": inp["цель"],
                "уровень_num": {"новичок": 0, "любитель": 1, "продвинутый": 2}.get(
                    inp["уровень_подготовки"], 0
                ),
            }
        )
        days_matches.append((req_days, out.get("тип_сплита")))

    df_X = pd.DataFrame(records)
    t0 = time.perf_counter()
    y_pred = [int(x) for x in model.predict(df_X).flatten().tolist()]
    total_ms = (time.perf_counter() - t0) * 1000

    matched_days = sum(
        1 for (req_days, _), pred_id in zip(days_matches, y_pred)
        if req_days > 0 and EXPECTED_DAYS_BY_SPLIT_ID.get(pred_id, 3) == req_days
    )
    n = max(len(test_data), 1)

    return {
        "model": "CatBoost Baseline v0",
        "sample_size": len(test_data),
        "macro_f1": round(float(f1_score(y_true, y_pred, average="macro", zero_division=0)), 4),
        "balanced_accuracy": round(float(balanced_accuracy_score(y_true, y_pred)), 4),
        "accuracy": round(float(accuracy_score(y_true, y_pred)), 4),
        "valid_json_rate": 1.0,
        "days_match_rate": round(matched_days / n, 4),
        "avg_latency_ms": round(total_ms / n, 2),
    }


def load_hf_model(model_path: str):
    tokenizer = AutoTokenizer.from_pretrained(model_path, trust_remote_code=True)
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token
    model = AutoModelForCausalLM.from_pretrained(
        model_path,
        torch_dtype=torch.float32,
        trust_remote_code=True,
    )
    if os.path.isdir(model_path) and os.path.exists(os.path.join(model_path, "adapter_config.json")):
        try:
            from peft import PeftModel

            base_model = AutoModelForCausalLM.from_pretrained(
                BASE_TRANSFORMER_NAME,
                torch_dtype=torch.float32,
                trust_remote_code=True,
            )
            model = PeftModel.from_pretrained(base_model, model_path)
        except Exception as e:
            print(f"Предупреждение при загрузке LoRA-адаптера: {e}")
    model.eval()
    return tokenizer, model


def evaluate_transformer(model_path: str, label: str, test_sample: list, max_new_tokens: int = 384) -> tuple[dict, list]:
    """Оценка трансформера на контрольном срезе теста."""
    tokenizer, model = load_hf_model(model_path)

    y_true, y_pred, latencies, examples = [], [], [], []
    valid_json_count = 0
    days_match_count = 0

    for item in test_sample:
        inp = item["input"]
        req_days = len(inp.get("дни_тренировок", []))
        y_true.append(SPLIT_MAP_RU_TO_ID.get(item["output"].get("тип_сплита"), 0))

        prompt = (
            f"Составь план тренировок в формате JSON (поля: тип_сплита, еженедельный_план). "
            f"Клиент: {inp['возраст']} лет, рост {inp.get('рост')} см, вес {inp.get('вес')} кг, "
            f"пол {inp['пол']}, активность: {inp.get('тип_активности')}, цель: {inp['цель']}, "
            f"уровень: {inp['уровень_подготовки']}, травмы: {inp['травмы_или_болезни']}, "
            f"дней в неделю: {req_days} ({', '.join(inp.get('дни_тренировок', []))}). JSON:"
        )
        messages = [{"role": "user", "content": prompt}]
        if getattr(tokenizer, "chat_template", None):
            text_in = tokenizer.apply_chat_template(messages, tokenize=False, add_generation_prompt=True)
        else:
            text_in = prompt
        inputs = tokenizer(text_in, return_tensors="pt")

        t0 = time.perf_counter()
        with torch.no_grad():
            out_ids = model.generate(
                **inputs,
                max_new_tokens=max_new_tokens,
                do_sample=False,
                pad_token_id=tokenizer.pad_token_id,
            )
        lat_ms = (time.perf_counter() - t0) * 1000
        latencies.append(lat_ms)

        gen_text = tokenizer.decode(out_ids[0][inputs["input_ids"].shape[1] :], skip_special_tokens=True)
        plan = extract_json_object(gen_text)

        if isinstance(plan, dict) and "тип_сплита" in plan and isinstance(plan.get("еженедельный_план"), list):
            valid_json_count += 1
            if len(plan["еженедельный_план"]) == req_days:
                days_match_count += 1
            pred_split = SPLIT_MAP_RU_TO_ID.get(plan["тип_сплита"], 0)
        else:
            plan = {"raw_error": gen_text[:200]}
            pred_split = 0

        y_pred.append(pred_split)
        if len(examples) < 3:
            examples.append({"input": inp, "generated": plan, "latency_ms": round(lat_ms, 1)})

    n = max(len(test_sample), 1)
    metrics = {
        "model": label,
        "sample_size": len(test_sample),
        "macro_f1": round(float(f1_score(y_true, y_pred, average="macro", zero_division=0)), 4),
        "balanced_accuracy": round(float(balanced_accuracy_score(y_true, y_pred)), 4),
        "valid_json_rate": round(valid_json_count / n, 4),
        "days_match_rate": round(days_match_count / n, 4),
        "p50_latency_ms": round(float(np.percentile(latencies, 50)), 2),
        "p95_latency_ms": round(float(np.percentile(latencies, 95)), 2),
    }
    return metrics, examples


def main():
    parser = argparse.ArgumentParser(description="KAN-81: оценка baseline и трансформеров на фиксированном test")
    parser.add_argument("--dataset", default=DATASET_PATH)
    parser.add_argument("--output", default="experiment_results.json")
    parser.add_argument("--sample-size", type=int, default=20)
    args = parser.parse_args()

    train_d, val_d, test_d = load_and_split_data(args.dataset)
    print(f"Split sizes -> Train: {len(train_d)}, Val: {len(val_d)}, Test: {len(test_d)}")

    cb_metrics = evaluate_catboost_baseline(test_d)
    sample_slice = test_d[: args.sample_size]
    tr_proto_metrics, proto_plans = evaluate_transformer(
        BASE_TRANSFORMER_NAME,
        f"Transformer Prototype ({BASE_TRANSFORMER_NAME})",
        sample_slice,
    )

    report_artifacts = {
        "seed": SEED,
        "dataset_splits": {"train": len(train_d), "val": len(val_d), "test": len(test_d)},
        "baseline_catboost": cb_metrics,
        "transformer_prototype": tr_proto_metrics,
        "sample_generations": proto_plans,
    }

    if os.path.isdir(FINETUNED_TRANSFORMER_DIR):
        tr_ft_metrics, ft_plans = evaluate_transformer(
            FINETUNED_TRANSFORMER_DIR,
            f"Transformer Fine-Tuned ({FINETUNED_TRANSFORMER_DIR})",
            sample_slice,
        )
        report_artifacts["transformer_finetuned"] = tr_ft_metrics
        report_artifacts["finetuned_sample_generations"] = ft_plans

    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(report_artifacts, f, ensure_ascii=False, indent=2)
    print(f"Готово! Результаты сохранены в {args.output}")


if __name__ == "__main__":
    main()