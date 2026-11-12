"""
Локальный ML-сервис для backend: POST /ai/user/chat → text/plain/user потоковый или статический ответ.
Возвращает 
"role": "assistant", - роль модели
"content": msg.content, - ответ модели
"program_ids": [] - program_ids если сработал пайплайн рекомендации он не пустой если нашло, если обычный чат то он всегда пустой
Запуск из папки ai:  python -m uvicorn chat.chat_ml_service:app --host 127.0.0.1 --port 8000
"""
import os
import json
from typing import Dict, Optional, List, Any
from dotenv import load_dotenv
from fastapi import FastAPI
from pydantic import BaseModel
from openai import OpenAI
from sqlalchemy import create_engine
from .logger_config import logger
from ai.chat.recommended_system.pipline import RecommendationPipline
from ai.chat.recommended_system.sql_worker import SqlFillterPlans
from ai.chat.recommended_system.model_worker import CrossEncoderRanker
import time 

load_dotenv()

# --- Инструменты (исправлена схема и свойства) ---
tools = [
    {
        "type": "function",
        "function": {
            "name": "search_workouts",
            "description": "Поиск и подбор программ тренировок под запрос пользователя",
            "parameters": {
                "type": "object",
                "properties": {
                    "target": {
                        "type": "string",
                        "description": "Целевая зона или цель: спина, похудение, ноги, разминка"
                    }
                },
                "required": ["target"]
            }
        }
    }
]

# --- Инициализация тяжелых ресурсов 1 РАЗ при старте ---
DATABASE_URL = "sqlite:///marketplace.db"
engine = create_engine(DATABASE_URL)
cross_ranker = CrossEncoderRanker(model_name="cross-encoder/mmarco-mMiniLMv2-L12-H384-v1")  # Создаем модель один раз

client = OpenAI(
    base_url="https://openrouter.ai/api/v1",
    api_key=os.getenv("API_KEY"),
)

app = FastAPI(title="sportik-ml-chat", version="0.1.0")


class ChatRequest(BaseModel):
    messages: List[Dict[str, Any]]
    user: dict
    plan: Optional[dict] = None


def build_system_prompt(user: dict, plan: Optional[dict]) -> str:
    return f"""Ты AI фитнес-ассистент.

Правила:
- Не давай медицинских диагнозов
- Не советуй опасные упражнения при наличии ограничений
- Отвечай кратко и структурировано
- Если пользователь новичок, не предлагай сложные упражнения
- Если возраст > 60, избегай высокоинтенсивных нагрузок
- Будь вежлив

Параметры пользователя: {user}
Текущий план: {plan}

Если пользователь просит подобрать/посоветовать тренировку — ОБЯЗАТЕЛЬНО вызови инструмент search_workouts.
Когда получишь список тренировок от инструмента, сравни их с жалобой пользователя, выбери ровно 3 лучшие и обязательно укажи их program_id."""


def get_marketplace_workout(user: dict, query: str) -> str:
    sql_work = SqlFillterPlans(user, engine)
    pipe = RecommendationPipline(ranker=cross_ranker, fillter=sql_work)
    market_plans = pipe.pipline(query)
    
    if market_plans is None or market_plans.empty:
        return json.dumps([], ensure_ascii=False)
        
    data = market_plans[["program_id", "description"]].to_dict(orient="records")
    return json.dumps(data, ensure_ascii=False)


@app.post("/ai/user/chat")
async def ai_chat(chat: ChatRequest) -> Dict[str, Any]:
    t_start = time.perf_counter()
    user_id = chat.user.get("user_id") or chat.user.get("id", "unknown")
    user_query = chat.messages[-1].get("content", "") if chat.messages else ""
    
    logger.info("Хэндлер чата принял данные | user_id: %s | запрос: '%s'", user_id, user_query)

    # Формируем контекст диалога
    conversation = [
        {"role": "system", "content": build_system_prompt(chat.user, chat.plan)}
    ] + chat.messages[-10:]  

    # Первый запрос к модели
    t_llm1 = time.perf_counter()
    response = client.chat.completions.create(
        model="qwen/qwen-2.5-72b-instruct",
        messages=conversation,
        tools=tools,
        tool_choice="auto"
    )
    msg = response.choices[0].message
    logger.info("Первый ответ модели получен за %.3f сек", time.perf_counter() - t_llm1)

    # Модель решила вызвать инструмент
    if msg.tool_calls:
        t_tool_start = time.perf_counter()
        tool_call = msg.tool_calls[0]
        
        call_args = json.loads(tool_call.function.arguments)
        search_query = call_args.get("target") or user_query

        logger.info("Модель вызвала инструмент '%s' с запросом: '%s'", tool_call.function.name, search_query)

        # Запускаем рекомендации
        t_pipe = time.perf_counter()
        data_for_model = get_marketplace_workout(chat.user, search_query)
        found_count = len(json.loads(data_for_model)) if data_for_model else 0
        logger.info("Пайплайн нашел %d программ за %.3f сек", found_count, time.perf_counter() - t_pipe)

        # Добавляем в контекст намерение модели и ответ тула
        conversation.append(msg)
        conversation.append({
            "role": "tool",
            "tool_call_id": tool_call.id,
            "name": tool_call.function.name,
            "content": data_for_model
        })

        logger.info("Модель начинает выбирать 3 лучшие тренировки")
        final_response = client.chat.completions.create(
            model="qwen/qwen-2.5-72b-instruct",
            messages=conversation
        )
        
        logger.info("Модель вернула ответ после tool за %.3f сек", time.perf_counter() - t_tool_start)
        logger.info("Запрос успешно обработан за суммарные %.3f сек", time.perf_counter() - t_start)

        selected_ids = [item["program_id"] for item in json.loads(data_for_model)[:3]] if data_for_model else []
        return {
            "role": "assistant",
            "content": final_response.choices[0].message.content,
            "program_ids": selected_ids
        }

    # Обычная беседа без тренировок
    logger.info("Обычный ответ без вызова рекомендаций | Суммарное время: %.3f сек", time.perf_counter() - t_start)
    return {
        "role": "assistant",
        "content": msg.content,
        "program_ids": []
    }