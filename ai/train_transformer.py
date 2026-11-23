import argparse
import json
import random
from pathlib import Path
import numpy as np
import torch
from sklearn.model_selection import train_test_split
from torch.utils.data import Dataset
from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
    DataCollatorForSeq2Seq,
    Trainer,
    TrainingArguments,
)

SEED = 42
random.seed(SEED)
np.random.seed(SEED)
torch.manual_seed(SEED)

REQUIRED_INPUT_FIELDS = [
    "возраст",
    "рост",
    "вес",
    "пол",
    "тип_активности",
    "травмы_или_болезни",
    "цель",
    "уровень_подготовки",
    "дни_тренировок",
]

SPLIT_NORMALIZATION = {
    "фулбади": "Фулбади",
    "фулбоди": "Фулбади",
    "full_body": "Фулбади",
    "верхниз": "ВерхНиз",
    "upper_lower": "ВерхНиз",
    "тяни-толкай": "Тяни-толкай",
    "push_pull": "Тяни-толкай",
    "пуш-пул-ноги": "Пуш-пул-ноги",
    "push_pull_legs": "Пуш-пул-ноги",
    "сплит": "Сплит",
    "split": "Сплит",
}


def validate_and_normalize_record(obj: dict) -> dict | None:
    """KAN-80: проверка входных полей и приведение целевого плана к единой структуре."""
    if not isinstance(obj, dict):
        return None
    inp = obj.get("input")
    out = obj.get("output")
    if not isinstance(inp, dict) or not isinstance(out, dict):
        return None

    for field in REQUIRED_INPUT_FIELDS:
        if field not in inp or inp[field] is None:
            return None

    days = inp.get("дни_тренировок")
    if not isinstance(days, list) or len(days) == 0:
        return None

    normalized_input = {
        "возраст": int(inp["возраст"]),
        "рост": float(inp["рост"]),
        "вес": float(inp["вес"]),
        "пол": str(inp["пол"]).strip(),
        "тип_активности": str(inp["тип_активности"]).strip(),
        "травмы_или_болезни": str(inp["травмы_или_болезни"]).strip(),
        "цель": str(inp["цель"]).strip(),
        "уровень_подготовки": str(inp["уровень_подготовки"]).strip(),
        "дни_тренировок": [str(d).strip() for d in days if str(d).strip()],
        "оборудование": inp.get("оборудование", []),
        "исключения": inp.get("исключения", []),
    }
    if not normalized_input["дни_тренировок"]:
        return None

    raw_split = str(out.get("тип_сплита", "Фулбади")).strip()
    split_name = SPLIT_NORMALIZATION.get(raw_split.lower(), raw_split if raw_split else "Фулбади")

    weekly_plan = out.get("еженедельный_план")
    if not isinstance(weekly_plan, list) or len(weekly_plan) == 0:
        return None

    req_len = len(normalized_input["дни_тренировок"])
    letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    clean_days = []
    for idx in range(req_len):
        src_day = weekly_plan[idx % len(weekly_plan)]
        if not isinstance(src_day, dict):
            return None
        exercises = src_day.get("упражнения")
        if not isinstance(exercises, list) or len(exercises) == 0:
            return None
        clean_ex = []
        for ex in exercises:
            if isinstance(ex, dict) and ex.get("группа"):
                grp = str(ex["группа"]).strip()
                sub = None if grp == "Пресс" else ex.get("подгруппа")
                clean_ex.append({"группа": grp, "подгруппа": sub})
        if not clean_ex:
            return None
        clean_days.append(
            {
                "день": letters[idx] if idx < len(letters) else str(idx + 1),
                "тип_дня": str(src_day.get("тип_дня", f"{split_name} {letters[idx]}")).strip(),
                "упражнения": clean_ex,
            }
        )

    normalized_output = {
        "тип_сплита": split_name,
        "еженедельный_план": clean_days,
    }
    return {"input": normalized_input, "output": normalized_output}


def load_clean_dataset(path: str) -> list[dict]:
    """Загрузка JSONL, фильтрация битых строк и дедупликация по профилю."""
    cleaned = []
    seen = set()
    with open(path, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                raw_obj = json.loads(line)
            except json.JSONDecodeError:
                continue
            norm = validate_and_normalize_record(raw_obj)
            if norm is None:
                continue
            sig = json.dumps(norm["input"], sort_keys=True, ensure_ascii=False)
            if sig not in seen:
                seen.add(sig)
                cleaned.append(norm)
    return cleaned


def format_prompt(tokenizer, inp: dict) -> str:
    days = inp["дни_тренировок"]
    user_msg = (
        "Составь план тренировок на неделю в строгом JSON с ключами "
        "'тип_сплита' и 'еженедельный_план' (ровно "
        f"{len(days)} дней: {', '.join(days)}). "
        "Каждый день содержит 'день', 'тип_дня', 'упражнения' ([{'группа': ..., 'подгруппа': ...}]).\n"
        f"Профиль: возраст={inp['возраст']}, рост={inp['рост']}, вес={inp['вес']}, "
        f"пол={inp['пол']}, активность={inp['тип_активности']}, "
        f"травмы={inp['травмы_или_болезни']}, цель={inp['цель']}, "
        f"уровень={inp['уровень_подготовки']}, дни={days}, "
        f"оборудование={inp.get('оборудование', [])}, исключения={inp.get('исключения', [])}."
    )
    messages = [{"role": "user", "content": user_msg}]
    if getattr(tokenizer, "chat_template", None):
        return tokenizer.apply_chat_template(messages, tokenize=False, add_generation_prompt=True)
    return user_msg + "\nJSON:\n"


class SFTPlanDataset(Dataset):
    def __init__(self, records: list[dict], tokenizer, max_len: int = 1536):
        self.records = records
        self.tokenizer = tokenizer
        self.max_len = max_len

    def __len__(self):
        return len(self.records)

    def __getitem__(self, idx):
        rec = self.records[idx]
        prompt = format_prompt(self.tokenizer, rec["input"])
        target = json.dumps(rec["output"], ensure_ascii=False, separators=(",", ":"))
        eos = self.tokenizer.eos_token or ""

        prompt_ids = self.tokenizer(prompt, add_special_tokens=False)["input_ids"]
        target_ids = self.tokenizer(target + eos, add_special_tokens=False)["input_ids"]

        input_ids = (prompt_ids + target_ids)[: self.max_len]
        labels = ([-100] * min(len(prompt_ids), self.max_len) + target_ids)[: self.max_len]
        attention_mask = [1] * len(input_ids)

        return {
            "input_ids": input_ids,
            "attention_mask": attention_mask,
            "labels": labels,
        }


def main():
    parser = argparse.ArgumentParser(description="KAN-80 & KAN-81: подготовка данных и дообучение трансформера")
    parser.add_argument("--dataset", default="dataset_2.jsonlines")
    parser.add_argument("--model-name", default="Qwen/Qwen2.5-0.5B-Instruct")
    parser.add_argument("--output-dir", default="models/qwen2.5-0.5b-workout-ft")
    parser.add_argument("--processed-out", default="data/processed_clean_dataset.jsonl")
    parser.add_argument("--epochs", type=int, default=2)
    parser.add_argument("--max-train-samples", type=int, default=100, help="0 для всех примеров")
    parser.add_argument("--lr", type=float, default=5e-5)
    args = parser.parse_args()

    records = load_clean_dataset(args.dataset)
    if not records:
        raise RuntimeError("После валидации не осталось ни одного валидного примера.")

    Path(args.processed_out).parent.mkdir(parents=True, exist_ok=True)
    with open(args.processed_out, "w", encoding="utf-8") as f:
        for r in records:
            f.write(json.dumps(r, ensure_ascii=False) + "\n")

    train_val, _ = train_test_split(records, test_size=0.15, random_state=SEED)
    train_data, val_data = train_test_split(train_val, test_size=0.15, random_state=SEED)

    if args.max_train_samples > 0:
        train_data = train_data[: args.max_train_samples]
        val_data = val_data[: min(len(val_data), 20)]

    print(f"Валидных записей: {len(records)} | Train: {len(train_data)} | Val: {len(val_data)}")

    tokenizer = AutoTokenizer.from_pretrained(args.model_name, trust_remote_code=True)
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token

    model = AutoModelForCausalLM.from_pretrained(
        args.model_name,
        torch_dtype=torch.float32,
        trust_remote_code=True,
    )

    try:
        from peft import LoraConfig, get_peft_model

        lora_cfg = LoraConfig(
            r=8,
            lora_alpha=16,
            target_modules=["q_proj", "v_proj"],
            lora_dropout=0.05,
            bias="none",
            task_type="CAUSAL_LM",
        )
        model = get_peft_model(model, lora_cfg)
        model.print_trainable_parameters()
    except ImportError:
        print("peft не установлен — выполняем стандартное дообучение весов.")

    train_ds = SFTPlanDataset(train_data, tokenizer)
    val_ds = SFTPlanDataset(val_data, tokenizer)
    collator = DataCollatorForSeq2Seq(tokenizer=tokenizer, padding=True, pad_to_multiple_of=8)

    Path(args.output_dir).mkdir(parents=True, exist_ok=True)
    training_args = TrainingArguments(
        output_dir=args.output_dir,
        num_train_epochs=args.epochs,
        per_device_train_batch_size=2,
        per_device_eval_batch_size=2,
        gradient_accumulation_steps=4,
        learning_rate=args.lr,
        logging_steps=10,
        save_strategy="epoch",
        eval_strategy="epoch" if len(val_ds) > 0 else "no",
        report_to="none",
        seed=SEED,
        use_cpu=not torch.cuda.is_available(),
    )

    trainer = Trainer(
        model=model,
        args=training_args,
        train_dataset=train_ds,
        eval_dataset=val_ds if len(val_ds) > 0 else None,
        data_collator=collator,
    )

    trainer.train()
    model.save_pretrained(args.output_dir)
    tokenizer.save_pretrained(args.output_dir)
    print(f"Модель сохранена в {args.output_dir}")


if __name__ == "__main__":
    main()