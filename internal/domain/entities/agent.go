package entities

import "time"

// Tier representa la complejidad de una tarea; decide qué agentes pueden atenderla
type Tier string

const (
	TierSimple  Tier = "simple"
	TierCode    Tier = "code"
	TierComplex Tier = "complex"
)

// Task es la unidad de trabajo que el orquestador delega a un agente
type Task struct {
	Prompt  string
	WorkDir string
	Tier    Tier
}

// AgentProfile describe cuándo y cuánto puede usarse un agente (una cuenta)
type AgentProfile struct {
	Name       string
	Tiers      []Tier
	Priority   int // menor = se prefiere antes (los más baratos primero)
	DailyLimit int // 0 = sin límite
	Enabled    bool
}

// Supports indica si el agente puede atender tareas de ese tier
func (p AgentProfile) Supports(tier Tier) bool {
	for _, t := range p.Tiers {
		if t == tier {
			return true
		}
	}
	return false
}

// AgentResult es la salida normalizada de cualquier CLI de agente
type AgentResult struct {
	Agent       string        `json:"agent"`
	Tier        Tier          `json:"tier"`
	Output      string        `json:"output"`
	ExitCode    int           `json:"exit_code"`
	Duration    time.Duration `json:"duration_ns"`
	CostUSD     float64       `json:"cost_usd,omitempty"`
	RateLimited bool          `json:"rate_limited"`
}

// AgentStatus resume el estado de un agente para el comando `tami agents`
type AgentStatus struct {
	Profile       AgentProfile
	Installed     bool
	RunsToday     int
	CooldownUntil time.Time
}
