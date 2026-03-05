from fastapi import FastAPI
from transformers import pipeline
import json

#Путь к модели
PATH = "models/model"


generate = pipeline("text-classification")

app = FastAPI()


@app.post("/plan/user")
async def genetate_plan(user: dict):
    return generate(str(user))