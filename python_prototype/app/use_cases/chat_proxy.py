import time
import uuid
import json
from app.domain.ports.memory_repository import IMemoryRepository
from app.domain.ports.llm_service import ILLMService
from app.domain.models.requests.chat_completion_request import ChatCompletionRequest
from app.domain.models.responses.chat_completion_response import ChatCompletionResponse
from app.domain.models.responses.choice import Choice
from app.domain.models.responses.choice_message import ChoiceMessage
from app.domain.models.responses.usage import Usage

class ChatProxyUseCase:
    """Orquestador principal Multi-Agente de TAMI."""
    
    def __init__(self, memory_repo: IMemoryRepository, llm_service: ILLMService):
        self.memory_repo = memory_repo
        self.llm_service = llm_service

    def execute(self, request: ChatCompletionRequest) -> ChatCompletionResponse:
        user_message = request.messages[-1].content
        
        # 1. El Clasificador decide el rumbo
        intent = self.llm_service.classify_intent(user_message)
        print(f"🚦 [AGENTE ENRUTADOR] Intención detectada: {intent}")

        final_text = ""

        # 2. Ruta A: Guardar Memoria
        if "GUARDAR" in intent:
            llm_response = self.llm_service.extract_memory_tool(user_message)
            if llm_response.tool_calls:
                for tool in llm_response.tool_calls:
                    args = json.loads(tool.arguments)
                    mem_id = f"mem-{uuid.uuid4().hex[:6]}"
                    self.memory_repo.save(mem_id, args["content"], args["category"])
                    final_text = f"🧠 [TAMI Memoria Actualizada]: Nueva regla en la categoría '{args['category']}'."
            else:
                final_text = "⚠️ Intenté guardar la regla pero el agente extractor falló."

        # 3. Ruta B: Responder Pregunta (RAG)
        else:
            context = self.memory_repo.retrieve_relevant(user_message)
            llm_response = self.llm_service.answer_question(user_message, context)
            
            # El fallback "or" protege contra respuestas vacías
            final_text = llm_response.content or "El modelo procesó la solicitud pero no generó texto."

        # 4. Formatear y retornar
        return ChatCompletionResponse(
            id=f"chatcmpl-{uuid.uuid4().hex[:12]}", 
            created=int(time.time()),
            model="mi-motor-multi-agente",
            choices=[
                Choice(
                    index=0, 
                    message=ChoiceMessage(role="assistant", content=final_text), 
                    finish_reason="stop"
                )
            ],
            usage=Usage(prompt_tokens=0, completion_tokens=0, total_tokens=0)
        )