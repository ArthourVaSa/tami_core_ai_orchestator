package usecases

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"tami-core/internal/domain/entities"

	repoPorts "tami-core/internal/domain/ports/repository"
	svcPorts "tami-core/internal/domain/ports/services"
)

var ErrNoAgentAvailable = errors.New("no hay agentes disponibles (todos sin instalar, deshabilitados, en cooldown o sin cupo)")

// OrchestratorUseCase decide qué agente (cuenta) atiende cada tarea para repartir el consumo de tokens
type OrchestratorUseCase struct {
	runners    map[string]svcPorts.IAgentRunner
	profiles   []entities.AgentProfile
	classifier svcPorts.ITaskClassifier
	usage      repoPorts.IUsageRepository
	cooldown   time.Duration
	now        func() time.Time
}

func NewOrchestratorUseCase(
	runners []svcPorts.IAgentRunner,
	profiles []entities.AgentProfile,
	classifier svcPorts.ITaskClassifier,
	usage repoPorts.IUsageRepository,
	cooldown time.Duration,
) *OrchestratorUseCase {
	byName := make(map[string]svcPorts.IAgentRunner, len(runners))
	for _, r := range runners {
		byName[r.Name()] = r
	}
	return &OrchestratorUseCase{
		runners:    byName,
		profiles:   profiles,
		classifier: classifier,
		usage:      usage,
		cooldown:   cooldown,
		now:        time.Now,
	}
}

// Execute clasifica la tarea, elige el mejor agente y hace fallback automático si una cuenta se queda sin cupo.
// forcedAgent (opcional) salta el enrutador y usa ese agente directamente.
func (uc *OrchestratorUseCase) Execute(ctx context.Context, prompt, workDir, forcedAgent string) (entities.AgentResult, error) {
	task := entities.Task{Prompt: prompt, WorkDir: workDir}

	var candidates []entities.AgentProfile
	if forcedAgent != "" {
		p, ok := uc.profile(forcedAgent)
		if !ok {
			return entities.AgentResult{}, fmt.Errorf("agente desconocido: %s", forcedAgent)
		}
		candidates = []entities.AgentProfile{p}
		task.Tier = entities.TierCode
	} else {
		task.Tier = uc.classify(prompt)
		candidates = uc.rank(task.Tier)
	}
	fmt.Printf("🚦 [ORQUESTADOR] Tier: %s | Candidatos: %s\n", task.Tier, names(candidates))

	if len(candidates) == 0 {
		return entities.AgentResult{}, ErrNoAgentAvailable
	}

	var lastErr error
	for _, p := range candidates {
		fmt.Printf("🤖 [ORQUESTADOR] Delegando a '%s'...\n", p.Name)
		result, err := uc.runners[p.Name].Run(ctx, task)
		result.Tier = task.Tier

		if result.RateLimited {
			until := uc.now().Add(uc.cooldown)
			_ = uc.usage.SetCooldown(p.Name, until)
			fmt.Printf("⏳ [ORQUESTADOR] '%s' sin cupo, en cooldown hasta %s. Probando el siguiente...\n", p.Name, until.Format("15:04"))
			lastErr = fmt.Errorf("%s: límite de uso alcanzado", p.Name)
			continue
		}

		_ = uc.usage.RecordRun(p.Name, err == nil, result.CostUSD)
		if err != nil {
			fmt.Printf("⚠️ [ORQUESTADOR] '%s' falló: %v. Probando el siguiente...\n", p.Name, err)
			lastErr = fmt.Errorf("%s: %w", p.Name, err)
			if ctx.Err() != nil {
				return result, lastErr
			}
			continue
		}
		return result, nil
	}
	return entities.AgentResult{}, fmt.Errorf("%w; último error: %v", ErrNoAgentAvailable, lastErr)
}

// Status devuelve el estado de todos los agentes configurados
func (uc *OrchestratorUseCase) Status() []entities.AgentStatus {
	out := make([]entities.AgentStatus, 0, len(uc.profiles))
	for _, p := range uc.profiles {
		out = append(out, entities.AgentStatus{
			Profile:       p,
			Installed:     uc.runners[p.Name] != nil && uc.runners[p.Name].IsInstalled(),
			RunsToday:     uc.usage.RunsToday(p.Name),
			CooldownUntil: uc.usage.CooldownUntil(p.Name),
		})
	}
	return out
}

// classify usa el modelo local; si falla, asumimos tarea de código (tier intermedio)
func (uc *OrchestratorUseCase) classify(prompt string) entities.Tier {
	if uc.classifier == nil {
		return entities.TierCode
	}
	tier, err := uc.classifier.ClassifyTask(prompt)
	if err != nil {
		fmt.Printf("⚠️ [ORQUESTADOR] Clasificador no disponible (%v), usando tier '%s'\n", err, entities.TierCode)
		return entities.TierCode
	}
	return tier
}

// rank ordena los agentes usables: primero los que soportan el tier (por prioridad y menor uso hoy),
// y al final el resto como último recurso para que la tarea nunca se quede sin atender.
func (uc *OrchestratorUseCase) rank(tier entities.Tier) []entities.AgentProfile {
	var preferred, fallback []entities.AgentProfile
	for _, p := range uc.profiles {
		if !uc.usable(p) {
			continue
		}
		if p.Supports(tier) {
			preferred = append(preferred, p)
		} else {
			fallback = append(fallback, p)
		}
	}
	uc.sortByCost(preferred)
	uc.sortByCost(fallback)
	return append(preferred, fallback...)
}

func (uc *OrchestratorUseCase) sortByCost(ps []entities.AgentProfile) {
	sort.SliceStable(ps, func(i, j int) bool {
		if ps[i].Priority != ps[j].Priority {
			return ps[i].Priority < ps[j].Priority
		}
		return uc.usage.RunsToday(ps[i].Name) < uc.usage.RunsToday(ps[j].Name)
	})
}

func (uc *OrchestratorUseCase) usable(p entities.AgentProfile) bool {
	r := uc.runners[p.Name]
	if !p.Enabled || r == nil || !r.IsInstalled() {
		return false
	}
	if uc.now().Before(uc.usage.CooldownUntil(p.Name)) {
		return false
	}
	return p.DailyLimit == 0 || uc.usage.RunsToday(p.Name) < p.DailyLimit
}

func (uc *OrchestratorUseCase) profile(name string) (entities.AgentProfile, bool) {
	for _, p := range uc.profiles {
		if p.Name == name && uc.runners[name] != nil {
			return p, true
		}
	}
	return entities.AgentProfile{}, false
}

func names(ps []entities.AgentProfile) string {
	n := make([]string, len(ps))
	for i, p := range ps {
		n[i] = p.Name
	}
	return strings.Join(n, " → ")
}
