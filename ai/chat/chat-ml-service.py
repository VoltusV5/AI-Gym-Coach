"""
От Go нужен вот такой формат сообщения:

{
"messages":[
    {"role": "user", "content": "Привет"},
    {"role": "assistant", "content": "Привет!"}
]
}
и так далее, мне нужны последнии 8 сообщений, 4 пары получается

"""
from fastapi import Body, FastAPI
from transformers import pipeline



pipe = pipeline("text-generation", model="microsoft/Phi-3-mini-4k-instruct")

def prompt(chat):
    prompt = "".join([i["role"]+ ": " + i["content"] + "\n" for i in chat["messages"]])
    prompt = "Ты ассистент. Отвечай КОРОТКО и по делу.\n\n" + prompt 
    print(prompt)
    return prompt
message = {"messages":[
    {"role": "Пользователь", "content": "Привет"},
]}
print(pipe(message))
app = FastAPI(title="sportik-ml-chat", version="0.1.0")

@app.post("/user/chat")
async def chat(chat):
    pass
