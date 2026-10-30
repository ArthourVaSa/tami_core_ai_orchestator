import json
from openai import OpenAI
from app.domain.ports.llm_service import ILLMService
from app.domain.models.core.llm_response import LLMResponse
from app.domain.models.core.tool_call_def import ToolCallDef
from app.config import global_settings  # <-- IMPORTANTE

class OllamaService(ILLMService):
    def __init__(self):
        self.client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
        self.tool_def = [{
            "type": "function",
            "function": {
                "name": "save_memory",
                "description": "Extrae la regla técnica para guardarla en memoria.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "category": {"type": "string"},
                        "content": {"type": "string"}
                    },
                    "required": ["category", "content"]
                }
            }
        }]

    def classify_intent(self, user_query: str) -> str:
        prompt = """Eres un clasificador de intenciones estricto.
Lee el mensaje del usuario y responde ÚNICAMENTE con la palabra 'GUARDAR' si está ordenando guardar una regla, contexto o decisión arquitectónica.
Responde ÚNICAMENTE con la palabra 'PREGUNTA' si está haciendo una consulta, dudando, o preguntando cómo hacer algo."""
        
        res = self.client.chat.completions.create(
            model=global_settings.active_model, # <-- LECTURA EN TIEMPO REAL
            messages=[{"role": "system", "content": prompt}, {"role": "user", "content": user_query}],
            temperature=0.0
        )
        return res.choices[0].message.content.strip().upper()

    def extract_memory_tool(self, user_query: str) -> LLMResponse:
        res = self.client.chat.completions.create(
            model=global_settings.active_model, # <-- LECTURA EN TIEMPO REAL
            messages=[{"role": "user", "content": f"Extrae los datos de esta regla: {user_query}"}],
            tools=self.tool_def,
            tool_choice={"type": "function", "function": {"name": "save_memory"}},
            temperature=0.1
        )
        message = res.choices[0].message
        parsed_tools = [ToolCallDef(name=tc.function.name, arguments=tc.function.arguments) for tc in message.tool_calls] if message.tool_calls else None
        return LLMResponse(content=message.content, tool_calls=parsed_tools)

    def answer_question(self, user_query: str, context: str) -> LLMResponse:
        prompt = f"Eres un Tech Lead de TAMI. Responde usando ESTRICTAMENTE este contexto:\n{context}"
        res = self.client.chat.completions.create(
            model=global_settings.active_model, # <-- LECTURA EN TIEMPO REAL
            messages=[{"role": "system", "content": prompt}, {"role": "user", "content": user_query}],
            temperature=0.3
        )
        return LLMResponse(content=res.choices[0].message.content, tool_calls=None)