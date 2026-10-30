from pydantic import BaseModel

class ChoiceMessage(BaseModel):
    role: str
    content: str