from typing import Any, AsyncIterator, Dict

from fastapi import Body, FastAPI
from fastapi.responses import StreamingResponse

app = FastAPI(title="sportik-ml-chat", version="0.1.0")


@app.get("/health")
async def health():
    return {"ok": True}


async def _stub_stream() -> AsyncIterator[bytes]:
    yield b'data: {"stub":true}\n\n'


@app.post("/ai/user/chat")
async def user_chat(_body: Dict[str, Any] = Body(default_factory=dict)):
    return StreamingResponse(
        _stub_stream(),
        media_type="text/event-stream",
        headers={"Cache-Control": "no-cache"},
    )
