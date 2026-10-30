package services

import (
	domainModels "tami-core/internal/domain/models/services"
)

type ILLMService interface {
	ClassifyIntent(userQuery string) string
	ExtractMemoryTool(userQuery string) domainModels.LLMResponse
	AnswerQuestion(userQuery string, context string) domainModels.LLMResponse
}
