from fastapi import APIRouter, Depends
from app.domain.models.requests.chat_completion_request import ChatCompletionRequest
from app.domain.models.responses.chat_completion_response import ChatCompletionResponse
from app.use_cases.chat_proxy import ChatProxyUseCase
from app.infrastructure.repositories.chroma_repo import ChromaRepository
from app.infrastructure.services.ollama_service import OllamaService
from app.config import global_settings
from pydantic import BaseModel

router = APIRouter()

# --- FÁBRICA DE DEPENDENCIAS (Inyección Dinámica) ---
def get_chat_proxy():
    # Instanciamos los adaptadores FRESCOS por cada petición que entra.
    # Así cada Hilo tiene su propia conexión segura a ChromaDB y a Ollama.
    memory_repo = ChromaRepository()
    llm_service = OllamaService() 
    return ChatProxyUseCase(memory_repo=memory_repo, llm_service=llm_service)


@router.post("/v1/chat/completions", response_model=ChatCompletionResponse)
async def chat_completions(
    request: ChatCompletionRequest,
    chat_proxy: ChatProxyUseCase = Depends(get_chat_proxy) # <-- Magia de FastAPI
):
    print(f"\n[📥 PROXY INTERCEPTÓ] Petición externa")
    print(f"🧠 Derivando tráfico localmente hacia -> {global_settings.active_model}")
    
    # Ejecutamos el caso de uso de forma segura
    response = chat_proxy.execute(request)
    return response

class ModelUpdateRequest(BaseModel):
    model_name: str

@router.post("/admin/config/model")
async def update_model(request: ModelUpdateRequest):
    global_settings.active_model = request.model_name
    print(f"🔧 [ADMIN] El servidor actualizó su motor de inferencia a: {request.model_name}")
    return {"status": "success", "active_model": global_settings.active_model}

@router.get("/admin/config/status")
async def get_status():
    return {"active_model": global_settings.active_model, "port": global_settings.port}