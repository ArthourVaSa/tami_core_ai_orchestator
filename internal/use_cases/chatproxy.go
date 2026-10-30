package usecases

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"tami-core/internal/domain/entities"
	"tami-core/internal/domain/models/request"
	"tami-core/internal/domain/models/response"

	repoPorts "tami-core/internal/domain/ports/repository"
	svcPorts "tami-core/internal/domain/ports/services"
)

// ChatProxyUseCase orquesta la lógica de negocio usando solo interfaces
type ChatProxyUseCase struct {
	memoryRepo repoPorts.IMemoryRepository
	llmService svcPorts.ILLMService
}

// NewChatProxyUseCase inyecta las dependencias al instanciar
func NewChatProxyUseCase(repo repoPorts.IMemoryRepository, llm svcPorts.ILLMService) *ChatProxyUseCase {
	return &ChatProxyUseCase{
		memoryRepo: repo,
		llmService: llm,
	}
}

// Execute es el punto de entrada para cada petición que llega desde Claude Code
func (uc *ChatProxyUseCase) Execute(req request.ChatCompletionRequest) response.ChatCompletionResponse {
	var userMessage string
	if len(req.Messages) > 0 {
		userMessage = req.Messages[len(req.Messages)-1].Content
	}

	finalText := ""

	// 1. Clasificamos la intención de la consulta
	intent := uc.llmService.ClassifyIntent(userMessage)
	fmt.Printf("🚦 [ENRUTADOR] Intención detectada: %s\n", intent)

	// 2. Ejecutamos la ruta correspondiente
	if strings.Contains(strings.ToUpper(intent), "GUARDAR") {
		// Ruta de extracción y guardado de memoria
		llmResp := uc.llmService.ExtractMemoryTool(userMessage)

		if len(llmResp.ToolCalls) > 0 {
			for _, tool := range llmResp.ToolCalls {
				var args map[string]string
				err := json.Unmarshal([]byte(tool.Arguments), &args)

				if err == nil {
					memID := fmt.Sprintf("mem-%d", time.Now().UnixMilli())
					uc.memoryRepo.Save(memID, args["content"], args["category"])
					finalText = fmt.Sprintf("🧠 [TAMI Memoria]: Regla guardada en '%s'.", args["category"])
				} else {
					finalText = "⚠️ Error interno al parsear la memoria."
				}
			}
		} else {
			finalText = "⚠️ El agente extractor no pudo procesar la regla."
		}
	} else {
		// Ruta RAG tradicional
		context := uc.memoryRepo.RetrieveRelevant(userMessage)
		llmResp := uc.llmService.AnswerQuestion(userMessage, context)

		if llmResp.Content != "" {
			finalText = llmResp.Content
		} else {
			finalText = "El modelo no generó una respuesta de texto válida."
		}
	}

	// 3. Empaquetamos la respuesta en el formato estándar HTTP
	responseMsg := entities.Message{
		Role:    "assistant",
		Content: finalText,
	}

	choice := response.Choice{
		Index: 0,
		Message: entities.Message{
			Role:    responseMsg.Role,
			Content: responseMsg.Content,
		},
		FinishReason: "stop",
	}

	return response.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "tami-multi-agent-core",
		Choices: []response.Choice{choice},
		Usage: response.Usage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}
}
