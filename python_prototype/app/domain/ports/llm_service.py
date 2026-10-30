from abc import ABC, abstractmethod
from app.domain.models.core.llm_response import LLMResponse

class ILLMService(ABC):
    
    @abstractmethod
    def classify_intent(self, user_query: str) -> str:
        """Agente 1: Analiza si el usuario quiere GUARDAR una regla o hacer una PREGUNTA."""
        pass

    @abstractmethod
    def extract_memory_tool(self, user_query: str) -> LLMResponse:
        """Agente 2 (Modo Escritura): Extrae los parámetros JSON para guardar en BD."""
        pass

    @abstractmethod
    def answer_question(self, user_query: str, context: str) -> LLMResponse:
        """Agente 3 (Modo Lectura): Responde la pregunta usando el contexto (SIN acceso a herramientas)."""
        pass