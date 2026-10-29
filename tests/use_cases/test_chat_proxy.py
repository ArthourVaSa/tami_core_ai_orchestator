import pytest
from app.use_cases.chat_proxy import ChatProxyUseCase
from app.domain.ports.memory_repository import IMemoryRepository
from app.domain.ports.llm_service import ILLMService
from app.domain.models.requests.chat_completion_request import ChatCompletionRequest
from app.domain.models.requests.message import Message
from app.domain.models.core.llm_response import LLMResponse
from app.domain.models.core.tool_call_def import ToolCallDef

# --- 1. MOCKS (Falsificamos la Infraestructura) ---

class MockMemoryRepository(IMemoryRepository):
    def __init__(self):
        self.saved_memories = []

    def save(self, memory_id: str, content: str, category: str) -> None:
        self.saved_memories.append({"id": memory_id, "content": content, "category": category})

    def retrieve_relevant(self, query: str, limit: int = 2) -> str:
        # Simulamos que encuentra una regla de arquitectura
        return "Regla: Usar Clean Architecture"


class MockLLMService(ILLMService):
    def __init__(self, will_use_tool: bool = False):
        self.will_use_tool = will_use_tool

    def generate_response(self, user_query: str, context: str) -> LLMResponse:
        if self.will_use_tool:
            # Simulamos que el LLM decide guardar una memoria
            return LLMResponse(
                content=None,
                tool_calls=[ToolCallDef(name="save_memory", arguments='{"category": "test", "content": "Guardar esto"}')]
            )
        # Simulamos una respuesta normal
        return LLMResponse(content="Respuesta normal del LLM", tool_calls=None)

# --- 2. LOS TESTS ---

def test_chat_proxy_returns_normal_response():
    """Prueba que el orquestador devuelve texto cuando el LLM decide hablar."""
    # Arrange
    memory_repo = MockMemoryRepository()
    llm_service = MockLLMService(will_use_tool=False)
    use_case = ChatProxyUseCase(memory_repo, llm_service)
    
    request = ChatCompletionRequest(
        model="gpt-4",
        messages=[Message(role="user", content="Hola")]
    )

    # Act
    response = use_case.execute(request)

    # Assert
    assert response.choices[0].message.content == "Respuesta normal del LLM"
    assert response.model == "mi-motor-hibrido-v1"


def test_chat_proxy_handles_tool_call():
    """Prueba que el orquestador guarda en la BD cuando el LLM usa la herramienta."""
    # Arrange
    memory_repo = MockMemoryRepository()
    llm_service = MockLLMService(will_use_tool=True)
    use_case = ChatProxyUseCase(memory_repo, llm_service)
    
    request = ChatCompletionRequest(
        model="gpt-4",
        messages=[Message(role="user", content="Guarda esta nueva regla")]
    )

    # Act
    response = use_case.execute(request)

    # Assert
    # 1. Verificamos que se llamó al repositorio para guardar
    assert len(memory_repo.saved_memories) == 1
    assert memory_repo.saved_memories[0]["content"] == "Guardar esto"
    
    # 2. Verificamos que se le responde algo al cliente CLI
    assert "Memoria Interna Actualizada" in response.choices[0].message.content