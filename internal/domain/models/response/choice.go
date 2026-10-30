package response

import "tami-core/internal/domain/entities"

type Choice struct {
	Index        int              `json:"index"`
	Message      entities.Message `json:"message"`
	FinishReason string           `json:"finish_reason"`
}
