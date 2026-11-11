package services

import (
	"context"

	"tami-core/internal/domain/entities"
)

// IAgentRunner abstrae cualquier CLI de agente (claude, gemini, codex, opencode, kimi...)
type IAgentRunner interface {
	Name() string
	IsInstalled() bool
	Run(ctx context.Context, task entities.Task) (entities.AgentResult, error)
}

// ITaskClassifier decide la complejidad de una tarea (idealmente con un modelo local y gratuito)
type ITaskClassifier interface {
	ClassifyTask(prompt string) (entities.Tier, error)
}
