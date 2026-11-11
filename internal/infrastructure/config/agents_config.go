package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"tami-core/internal/domain/entities"
)

// AgentConfig define cómo invocar un CLI de agente. "{prompt}" en Args se sustituye por la tarea.
type AgentConfig struct {
	Name       string            `json:"name"`
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	Env        map[string]string `json:"env,omitempty"`    // p.ej. CLAUDE_CONFIG_DIR para usar otra cuenta
	Output     string            `json:"output,omitempty"` // "text" (default) | "claude-json"
	Tiers      []entities.Tier   `json:"tiers"`
	Priority   int               `json:"priority"`
	DailyLimit int               `json:"daily_limit"`
	Enabled    bool              `json:"enabled"`
}

type OrchestratorConfig struct {
	OllamaURL         string        `json:"ollama_url"`
	ClassifierModel   string        `json:"classifier_model"`
	CooldownMinutes   int           `json:"cooldown_minutes"`
	TimeoutMinutes    int           `json:"timeout_minutes"`
	RateLimitPatterns []string      `json:"rate_limit_patterns"`
	Agents            []AgentConfig `json:"agents"`
}

func (a AgentConfig) Profile() entities.AgentProfile {
	return entities.AgentProfile{
		Name:       a.Name,
		Tiers:      a.Tiers,
		Priority:   a.Priority,
		DailyLimit: a.DailyLimit,
		Enabled:    a.Enabled,
	}
}

// Default: los agentes más baratos atienden lo simple; Claude se reserva para lo complejo.
// Los flags de cada CLI cambian con sus versiones: ajústalos en ~/.tami/agents.json.
func Default() OrchestratorConfig {
	return OrchestratorConfig{
		OllamaURL:       "http://localhost:11434",
		ClassifierModel: "llama3.1:latest",
		CooldownMinutes: 60,
		TimeoutMinutes:  30,
		RateLimitPatterns: []string{
			"rate limit", "rate_limit", "usage limit", "limit reached",
			"quota", "resource_exhausted", "too many requests", "429",
		},
		Agents: []AgentConfig{
			{
				Name: "gemini", Command: "gemini", Args: []string{"--yolo", "-p", "{prompt}"},
				Tiers: []entities.Tier{entities.TierSimple, entities.TierCode}, Priority: 1, Enabled: true,
			},
			{
				Name: "opencode", Command: "opencode", Args: []string{"run", "{prompt}"},
				Tiers: []entities.Tier{entities.TierSimple, entities.TierCode}, Priority: 2, Enabled: true,
			},
			{
				Name: "kimi", Command: "kimi", Args: []string{"--print", "--command", "{prompt}"},
				Tiers: []entities.Tier{entities.TierSimple, entities.TierCode}, Priority: 2, Enabled: true,
			},
			{
				Name: "codex", Command: "codex", Args: []string{"exec", "--full-auto", "{prompt}"},
				Tiers: []entities.Tier{entities.TierCode, entities.TierComplex}, Priority: 3, Enabled: true,
			},
			{
				Name: "claude", Command: "claude",
				Args:   []string{"-p", "{prompt}", "--output-format", "json", "--permission-mode", "acceptEdits"},
				Output: "claude-json",
				Tiers:  []entities.Tier{entities.TierComplex}, Priority: 4, Enabled: true,
			},
		},
	}
}

// DefaultPath devuelve ~/.tami/agents.json
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "agents.json"
	}
	return filepath.Join(home, ".tami", "agents.json")
}

// Load lee la configuración; si no existe, escribe la de por defecto para que el usuario la edite
func Load(path string) (OrchestratorConfig, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg := Default()
		if err := write(path, cfg); err != nil {
			return cfg, fmt.Errorf("no se pudo crear %s: %w", path, err)
		}
		fmt.Printf("📝 [CONFIG] Configuración por defecto creada en %s\n", path)
		return cfg, nil
	}
	if err != nil {
		return OrchestratorConfig{}, err
	}

	cfg := Default()
	cfg.Agents = nil
	if err := json.Unmarshal(data, &cfg); err != nil {
		return OrchestratorConfig{}, fmt.Errorf("JSON inválido en %s: %w", path, err)
	}
	return cfg, nil
}

func write(path string, cfg OrchestratorConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
