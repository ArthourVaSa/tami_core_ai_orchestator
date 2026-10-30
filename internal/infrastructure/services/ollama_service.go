package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	svcModels "tami-core/internal/domain/models/services"
	"tami-core/internal/infrastructure/dtos/request"
	"tami-core/internal/infrastructure/dtos/response"
)

type OllamaService struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaService(baseURL string, model string) *OllamaService {
	return &OllamaService{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

// sendPrompt maneja exclusivamente la fontanería HTTP
func (s *OllamaService) sendPrompt(userQuery string, systemPrompt string) string {
	reqBody := request.OllamaGenerateRequest{
		Model:  s.model,
		Prompt: userQuery,
		System: systemPrompt,
		Stream: false,
	}

	jsonData, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/api/generate", s.baseURL)

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error de red: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var ollamaResp response.OllamaGenerateResponse
	json.Unmarshal(bodyBytes, &ollamaResp)

	return strings.TrimSpace(ollamaResp.Response)
}

// --- Implementación de ILLMService ---

func (s *OllamaService) ClassifyIntent(userQuery string) string {
	systemPrompt := `Eres un enrutador estricto. Clasifica la petición del usuario en: 
1. "GUARDAR": Si el usuario te da una regla o preferencia.
2. "CONSULTAR": Si hace una pregunta general.
Responde ÚNICAMENTE con la palabra GUARDAR o CONSULTAR.`

	return s.sendPrompt(userQuery, systemPrompt)
}

func (s *OllamaService) ExtractMemoryTool(userQuery string) svcModels.LLMResponse {
	systemPrompt := `Extrae la regla del usuario y devuélvela EXACTAMENTE en este JSON:
{"content": "regla", "category": "categoria"}`

	rawJSON := s.sendPrompt(userQuery, systemPrompt)
	tool := svcModels.ToolCallDef{Name: "save_memory", Arguments: rawJSON}

	return svcModels.LLMResponse{Content: "", ToolCalls: []svcModels.ToolCallDef{tool}}
}

func (s *OllamaService) AnswerQuestion(userQuery string, context string) svcModels.LLMResponse {
	systemPrompt := fmt.Sprintf(`Responde la consulta utilizando este contexto de memoria: 
---
%s
---`, context)

	answer := s.sendPrompt(userQuery, systemPrompt)
	return svcModels.LLMResponse{Content: answer, ToolCalls: nil}
}
