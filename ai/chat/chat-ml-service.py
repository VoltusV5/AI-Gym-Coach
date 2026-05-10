"""
Локальный ML-сервис для backend: POST /ai/user/chat → JSON ответ {"role":"assistant", "content": "Что то на нейронском"}.
Запуск из папки ai:  python -m uvicorn chat.chat-ml-service:app --host 127.0.0.1 --port 5051
"""


from fastapi import Body, FastAPI
from pydantic import BaseModel
from cerebras.cloud.sdk import Cerebras
from dotenv import load_dotenv
from typing import Any, Dict, Optional
import os

def prompt(chat: list[dict], user,plan) -> list[dict[str,str]]:
    """
    Функция создающая промпт для нейроки и задающая её поведение.
    """
    system =  [{
        "role": "system",
        "content": f"""
        Ты AI фитнес-ассистент.

        Правила:
        - Не давай медицинских диагнозов
        - Не советуй опасные упражнения при наличии ограничений
        - Отвечай кратко и структурировано
        - Если пользователь новичок, не предлагай сложные упражнения
        - Если возраст > 60, избегай высокоинтенсивных нагрузок
        - Будь вежлив

        Вот параметры человека: {user}
        Вот тренировочный план его: {plan}

        """
    }]

    message = system + chat[:10] + [{"role": "assistant", "content": ""}]
    return message




class ChatRequest(BaseModel):
    messages: list
    user: dict
    plan: Optional[dict] = None


load_dotenv()
client = Cerebras(
      api_key=os.getenv("CEREBRAS_API_KEY")
  )


app = FastAPI(title="sportik-ml-chat", version="0.1.0")

@app.post("/ai/user/chat")
async def ai_chat(chat:ChatRequest) -> Dict[str,str]:
    completion = client.chat.completions.create(
      messages=prompt(chat.messages, chat.user, chat.plan),
      model="llama3.1-8b",
      max_completion_tokens=1024,
      temperature=0.2,
      top_p=1,
      stream=False
    )
    
    return {"role":"assistant", "content": completion.choices[0].message.content}


