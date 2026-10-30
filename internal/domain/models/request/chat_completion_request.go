package request

import "tami-core/internal/domain/entities"

type ChatCompletionRequest struct {
	Model    string             `json:"model"`
	Messages []entities.Message `json:"messages"`
}
