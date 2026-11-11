package repositories

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type agentUsage struct {
	Day           string    `json:"day"` // YYYY-MM-DD; los contadores se reinician al cambiar de día
	Runs          int       `json:"runs"`
	Failures      int       `json:"failures"`
	CostUSD       float64   `json:"cost_usd"`
	CooldownUntil time.Time `json:"cooldown_until"`
}

// JSONUsageRepo implementa IUsageRepository persistiendo en un archivo JSON local (~/.tami/usage.json)
type JSONUsageRepo struct {
	path string
	mu   sync.Mutex
	data map[string]*agentUsage
	now  func() time.Time
}

func NewJSONUsageRepo(path string) (*JSONUsageRepo, error) {
	r := &JSONUsageRepo{path: path, data: map[string]*agentUsage{}, now: time.Now}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return r, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &r.data); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *JSONUsageRepo) RunsToday(agent string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.get(agent).Runs
}

func (r *JSONUsageRepo) RecordRun(agent string, success bool, costUSD float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u := r.get(agent)
	u.Runs++
	if !success {
		u.Failures++
	}
	u.CostUSD += costUSD
	return r.flush()
}

func (r *JSONUsageRepo) CooldownUntil(agent string) time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.get(agent).CooldownUntil
}

func (r *JSONUsageRepo) SetCooldown(agent string, until time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.get(agent).CooldownUntil = until
	return r.flush()
}

// get devuelve el registro del agente, reiniciando contadores si es un día nuevo (requiere mu)
func (r *JSONUsageRepo) get(agent string) *agentUsage {
	today := r.now().Format("2006-01-02")
	u, ok := r.data[agent]
	if !ok {
		u = &agentUsage{Day: today}
		r.data[agent] = u
	}
	if u.Day != today {
		*u = agentUsage{Day: today, CooldownUntil: u.CooldownUntil}
	}
	return u
}

func (r *JSONUsageRepo) flush() error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, raw, 0o644)
}
