package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"tami-core/internal/domain/entities"
	svcPorts "tami-core/internal/domain/ports/services"
)

type fakeRunner struct {
	name        string
	installed   bool
	rateLimited bool
	err         error
	calls       int
}

func (f *fakeRunner) Name() string      { return f.name }
func (f *fakeRunner) IsInstalled() bool { return f.installed }
func (f *fakeRunner) Run(_ context.Context, t entities.Task) (entities.AgentResult, error) {
	f.calls++
	if f.rateLimited {
		return entities.AgentResult{Agent: f.name, RateLimited: true}, errors.New("429")
	}
	return entities.AgentResult{Agent: f.name, Output: "ok"}, f.err
}

type fakeClassifier struct {
	tier entities.Tier
	err  error
}

func (f fakeClassifier) ClassifyTask(string) (entities.Tier, error) { return f.tier, f.err }

type fakeUsage struct {
	runs     map[string]int
	cooldown map[string]time.Time
}

func newFakeUsage() *fakeUsage {
	return &fakeUsage{runs: map[string]int{}, cooldown: map[string]time.Time{}}
}
func (f *fakeUsage) RunsToday(a string) int { return f.runs[a] }
func (f *fakeUsage) RecordRun(a string, _ bool, _ float64) error {
	f.runs[a]++
	return nil
}
func (f *fakeUsage) CooldownUntil(a string) time.Time { return f.cooldown[a] }
func (f *fakeUsage) SetCooldown(a string, t time.Time) error {
	f.cooldown[a] = t
	return nil
}

func profile(name string, prio, limit int, tiers ...entities.Tier) entities.AgentProfile {
	return entities.AgentProfile{Name: name, Priority: prio, DailyLimit: limit, Tiers: tiers, Enabled: true}
}

func TestOrchestratorExecute(t *testing.T) {
	cheapTiers := []entities.Tier{entities.TierSimple, entities.TierCode}

	tests := []struct {
		name      string
		tier      entities.Tier
		classErr  error
		runners   []*fakeRunner
		profiles  []entities.AgentProfile
		seedRuns  map[string]int
		seedCool  map[string]time.Time
		wantAgent string
		wantErr   bool
		wantCool  string // agente que debe acabar en cooldown
	}{
		{
			name:      "tarea simple va al agente más barato",
			tier:      entities.TierSimple,
			runners:   []*fakeRunner{{name: "gemini", installed: true}, {name: "claude", installed: true}},
			profiles:  []entities.AgentProfile{profile("claude", 4, 0, entities.TierComplex), profile("gemini", 1, 0, cheapTiers...)},
			wantAgent: "gemini",
		},
		{
			name:      "tarea compleja se reserva para claude",
			tier:      entities.TierComplex,
			runners:   []*fakeRunner{{name: "gemini", installed: true}, {name: "claude", installed: true}},
			profiles:  []entities.AgentProfile{profile("gemini", 1, 0, cheapTiers...), profile("claude", 4, 0, entities.TierComplex)},
			wantAgent: "claude",
		},
		{
			name:      "rate limit pone cooldown y hace fallback al siguiente",
			tier:      entities.TierSimple,
			runners:   []*fakeRunner{{name: "gemini", installed: true, rateLimited: true}, {name: "kimi", installed: true}},
			profiles:  []entities.AgentProfile{profile("gemini", 1, 0, cheapTiers...), profile("kimi", 2, 0, cheapTiers...)},
			wantAgent: "kimi",
			wantCool:  "gemini",
		},
		{
			name:      "sin agente del tier usa cualquier otro como último recurso",
			tier:      entities.TierSimple,
			runners:   []*fakeRunner{{name: "gemini", installed: false}, {name: "claude", installed: true}},
			profiles:  []entities.AgentProfile{profile("gemini", 1, 0, cheapTiers...), profile("claude", 4, 0, entities.TierComplex)},
			wantAgent: "claude",
		},
		{
			name:      "límite diario alcanzado salta al siguiente",
			tier:      entities.TierCode,
			runners:   []*fakeRunner{{name: "gemini", installed: true}, {name: "opencode", installed: true}},
			profiles:  []entities.AgentProfile{profile("gemini", 1, 5, cheapTiers...), profile("opencode", 2, 0, cheapTiers...)},
			seedRuns:  map[string]int{"gemini": 5},
			wantAgent: "opencode",
		},
		{
			name:      "misma prioridad reparte hacia el menos usado hoy",
			tier:      entities.TierCode,
			runners:   []*fakeRunner{{name: "kimi", installed: true}, {name: "opencode", installed: true}},
			profiles:  []entities.AgentProfile{profile("kimi", 2, 0, cheapTiers...), profile("opencode", 2, 0, cheapTiers...)},
			seedRuns:  map[string]int{"kimi": 3, "opencode": 1},
			wantAgent: "opencode",
		},
		{
			name:      "agente en cooldown se ignora",
			tier:      entities.TierCode,
			runners:   []*fakeRunner{{name: "gemini", installed: true}, {name: "codex", installed: true}},
			profiles:  []entities.AgentProfile{profile("gemini", 1, 0, cheapTiers...), profile("codex", 3, 0, entities.TierCode)},
			seedCool:  map[string]time.Time{"gemini": time.Now().Add(time.Hour)},
			wantAgent: "codex",
		},
		{
			name:      "clasificador caído asume tier code",
			classErr:  errors.New("ollama down"),
			runners:   []*fakeRunner{{name: "claude", installed: true}, {name: "codex", installed: true}},
			profiles:  []entities.AgentProfile{profile("claude", 4, 0, entities.TierComplex), profile("codex", 3, 0, entities.TierCode)},
			wantAgent: "codex",
		},
		{
			name:     "todos fallan devuelve error",
			tier:     entities.TierCode,
			runners:  []*fakeRunner{{name: "gemini", installed: true, err: errors.New("boom")}},
			profiles: []entities.AgentProfile{profile("gemini", 1, 0, cheapTiers...)},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := newFakeUsage()
			for k, v := range tt.seedRuns {
				usage.runs[k] = v
			}
			for k, v := range tt.seedCool {
				usage.cooldown[k] = v
			}
			runners := make([]svcPorts.IAgentRunner, len(tt.runners))
			for i, r := range tt.runners {
				runners[i] = r
			}

			uc := NewOrchestratorUseCase(runners, tt.profiles, fakeClassifier{tt.tier, tt.classErr}, usage, time.Hour)
			result, err := uc.Execute(context.Background(), "tarea", ".", "")

			if tt.wantErr {
				if !errors.Is(err, ErrNoAgentAvailable) {
					t.Fatalf("esperaba ErrNoAgentAvailable, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if result.Agent != tt.wantAgent {
				t.Errorf("agente = %s, want %s", result.Agent, tt.wantAgent)
			}
			if tt.wantCool != "" && !usage.cooldown[tt.wantCool].After(time.Now()) {
				t.Errorf("%s debería estar en cooldown", tt.wantCool)
			}
		})
	}
}

func TestOrchestratorForcedAgentSkipsClassifier(t *testing.T) {
	gemini := &fakeRunner{name: "gemini", installed: true}
	claude := &fakeRunner{name: "claude", installed: true}
	uc := NewOrchestratorUseCase(
		[]svcPorts.IAgentRunner{gemini, claude},
		[]entities.AgentProfile{profile("gemini", 1, 0, entities.TierSimple), profile("claude", 4, 0, entities.TierComplex)},
		fakeClassifier{tier: entities.TierSimple}, newFakeUsage(), time.Hour,
	)

	result, err := uc.Execute(context.Background(), "tarea", ".", "claude")
	if err != nil || result.Agent != "claude" || gemini.calls != 0 {
		t.Fatalf("esperaba claude forzado sin tocar gemini; got agent=%s err=%v gemini.calls=%d", result.Agent, err, gemini.calls)
	}

	if _, err := uc.Execute(context.Background(), "tarea", ".", "nope"); err == nil {
		t.Fatal("esperaba error con agente desconocido")
	}
}
