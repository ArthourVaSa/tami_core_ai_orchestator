from pydantic import BaseModel
from typing import List, Optional
from .tool_call_def import ToolCallDef

class LLMResponse(BaseModel):
    content: Optional[str] = None
    tool_calls: Optional[List[ToolCallDef]] = None