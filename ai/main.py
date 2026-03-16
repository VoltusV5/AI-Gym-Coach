from fastapi import FastAPI
from transformers import pipeline
from services.utils import translate_ru_json, generate
import json

#Путь к модели
PATH = "models/model"

app = FastAPI()



@app.post("/plan/user")
async def genetate_plan(user: dict):
    print(translate_ru_json(user))
    return generate