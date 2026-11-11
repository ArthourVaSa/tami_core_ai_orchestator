package services

import (
	"errors"
	"strings"

	"tami-core/internal/domain/entities"
)

// ClassifyTask implementa ITaskClassifier: el modelo local (gratis) decide la complejidad
// para no gastar tokens de las cuentas de pago solo en enrutar.
func (s *OllamaService) ClassifyTask(prompt string) (entities.Tier, error) {
	systemPrompt := `Eres un clasificador estricto de tareas para un orquestador de agentes de programación.
Clasifica la tarea del usuario en:
1. "SIMPLE": preguntas, explicaciones, resúmenes o cambios triviales de una línea.
2. "CODE": implementar o arreglar código acotado (una función, un archivo, un test).
3. "COMPLEX": arquitectura, refactors grandes, debugging difícil o cambios en muchos archivos.
Responde ÚNICAMENTE con la palabra SIMPLE, CODE o COMPLEX.`

	answer := strings.ToUpper(s.sendPrompt(prompt, systemPrompt))
	switch {
	case answer == "":
		return "", errors.New("ollama no respondió")
	case strings.Contains(answer, "COMPLEX"):
		return entities.TierComplex, nil
	case strings.Contains(answer, "SIMPLE"):
		return entities.TierSimple, nil
	default:
		return entities.TierCode, nil
	}
}
