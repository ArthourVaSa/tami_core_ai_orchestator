from pydantic import BaseModel
from typing import List, Optional
from .message import Message

class ChatCompletionRequest(BaseModel):
    model: str
    messages: List[Message]
    temperature: Optional[float] = 0.7
    stream: Optional[bool] = False