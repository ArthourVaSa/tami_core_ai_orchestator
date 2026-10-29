from pydantic import BaseModel
from typing import List
from .choice import Choice
from .usage import Usage

class ChatCompletionResponse(BaseModel):
    id: str
    object: str = "chat.completion"
    created: int
    model: str
    choices: List[Choice]
    usage: Usage