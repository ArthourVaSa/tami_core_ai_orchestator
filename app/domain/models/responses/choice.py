from pydantic import BaseModel
from .choice_message import ChoiceMessage

class Choice(BaseModel):
    index: int
    message: ChoiceMessage
    finish_reason: str