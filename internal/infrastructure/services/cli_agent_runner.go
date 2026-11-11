package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"tami-core/internal/domain/entities"
	"tami-core/internal/infrastructure/config"
)

// CLIAgentRunner implementa IAgentRunner ejecutando cualquier CLI de agente en modo no interactivo
type CLIAgentRunner struct {
	cfg               config.AgentConfig
	rateLimitPatterns []string
	timeout           time.Duration
}

func NewCLIAgentRunner(cfg config.AgentConfig, rateLimitPatterns []string, timeout time.Duration) *CLIAgentRunner {
	return &CLIAgentRunner{cfg: cfg, rateLimitPatterns: rateLimitPatterns, timeout: timeout}
}

func (r *CLIAgentRunner) Name() string { return r.cfg.Name }

func (r *CLIAgentRunner) IsInstalled() bool {
	_, err := exec.LookPath(r.cfg.Command)
	return err == nil
}

func (r *CLIAgentRunner) Run(ctx context.Context, task entities.Task) (entities.AgentResult, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Sin shell: el prompt viaja como argumento literal, no hay riesgo de inyección
	cmd := exec.CommandContext(ctx, r.cfg.Command, r.buildArgs(task.Prompt)...)
	cmd.Dir = task.WorkDir
	cmd.Env = os.Environ()
	for k, v := range r.cfg.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, os.ExpandEnv(v)))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()

	result := entities.AgentResult{
		Agent:    r.cfg.Name,
		Output:   strings.TrimSpace(stdout.String()),
		Duration: time.Since(start),
	}
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
	}

	isError := runErr != nil
	if r.cfg.Output == "claude-json" {
		isError = r.parseClaudeJSON(&result) || isError
	}

	if isError {
		result.RateLimited = r.looksRateLimited(result.Output + "\n" + stderr.String())
		if ctx.Err() == context.DeadlineExceeded {
			return result, fmt.Errorf("timeout tras %s", r.timeout)
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = result.Output
		}
		return result, fmt.Errorf("exit %d: %s", result.ExitCode, truncate(msg, 500))
	}
	return result, nil
}

func (r *CLIAgentRunner) buildArgs(prompt string) []string {
	args := make([]string, len(r.cfg.Args))
	for i, a := range r.cfg.Args {
		args[i] = strings.ReplaceAll(a, "{prompt}", prompt)
	}
	return args
}

// parseClaudeJSON extrae el texto y el coste de `claude -p --output-format json`; devuelve si fue error
func (r *CLIAgentRunner) parseClaudeJSON(result *entities.AgentResult) bool {
	var out struct {
		Result       string  `json:"result"`
		IsError      bool    `json:"is_error"`
		TotalCostUSD float64 `json:"total_cost_usd"`
	}
	if err := json.Unmarshal([]byte(result.Output), &out); err != nil {
		return false
	}
	result.Output = out.Result
	result.CostUSD = out.TotalCostUSD
	return out.IsError
}

func (r *CLIAgentRunner) looksRateLimited(text string) bool {
	lower := strings.ToLower(text)
	for _, p := range r.rateLimitPatterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
