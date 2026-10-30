from pydantic import BaseModel

class ToolCallDef(BaseModel):
    name: str
    arguments: str  # JSON string