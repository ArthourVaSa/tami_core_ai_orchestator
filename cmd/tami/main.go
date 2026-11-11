package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	// Ajusta estos imports según el nombre real de tu módulo en go.mod (ej. "tami-core/...")
	"tami-core/internal/domain/entities"
	"tami-core/internal/domain/models/request"
	svcPorts "tami-core/internal/domain/ports/services"
	"tami-core/internal/infrastructure/config"
	"tami-core/internal/infrastructure/repositories"
	"tami-core/internal/infrastructure/services"
	usecases "tami-core/internal/use_cases"
)

const usage = `TAMI - Orquestador multi-agente

Uso:
  tami serve                        Arranca el gateway HTTP (por defecto)
  tami run [-agent X] [-dir D] "tarea"   Delega una tarea al agente más barato disponible
  tami agents                       Muestra agentes, instalación, uso de hoy y cooldowns

Config: ~/.tami/agents.json (se crea al primer uso; flag global -config para otra ruta)
`

func main() {
	configPath := flag.String("config", config.DefaultPath(), "ruta a agents.json")
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()

	cmd := flag.Arg(0)
	if cmd == "" {
		cmd = "serve"
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	// ==========================================
	// 1. INYECCIÓN DE DEPENDENCIAS (Wiring)
	// ==========================================
	llmService := services.NewOllamaService(cfg.OllamaURL, cfg.ClassifierModel)
	memoryRepo := repositories.NewChromaRepo("http://localhost:8000", "tami_memory")
	usageRepo, err := repositories.NewJSONUsageRepo(filepath.Join(filepath.Dir(*configPath), "usage.json"))
	if err != nil {
		log.Fatalf("❌ No se pudo leer usage.json: %v", err)
	}

	timeout := time.Duration(cfg.TimeoutMinutes) * time.Minute
	runners := make([]svcPorts.IAgentRunner, 0, len(cfg.Agents))
	for _, a := range cfg.Agents {
		runners = append(runners, services.NewCLIAgentRunner(a, cfg.RateLimitPatterns, timeout))
	}
	profiles := make([]entities.AgentProfile, 0, len(cfg.Agents))
	for _, a := range cfg.Agents {
		profiles = append(profiles, a.Profile())
	}

	orchestrator := usecases.NewOrchestratorUseCase(
		runners, profiles, llmService, usageRepo, time.Duration(cfg.CooldownMinutes)*time.Minute,
	)
	chatProxyUseCase := usecases.NewChatProxyUseCase(memoryRepo, llmService)

	switch cmd {
	case "serve":
		serve(chatProxyUseCase, orchestrator)
	case "run":
		runTask(orchestrator, flag.Args()[1:])
	case "agents":
		printAgents(orchestrator)
	default:
		flag.Usage()
		os.Exit(2)
	}
}

func runTask(orchestrator *usecases.OrchestratorUseCase, args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	agent := fs.String("agent", "", "forzar un agente concreto (salta el enrutador)")
	dir := fs.String("dir", ".", "directorio de trabajo para el agente")
	_ = fs.Parse(args)

	prompt := strings.Join(fs.Args(), " ")
	if prompt == "" {
		log.Fatal(`❌ Falta la tarea: tami run "explica este repo"`)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	result, err := orchestrator.Execute(ctx, prompt, *dir, *agent)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}
	fmt.Printf("✅ [%s | tier %s | %s", result.Agent, result.Tier, result.Duration.Round(time.Second))
	if result.CostUSD > 0 {
		fmt.Printf(" | $%.4f", result.CostUSD)
	}
	fmt.Printf("]\n\n%s\n", result.Output)
}

func printAgents(orchestrator *usecases.OrchestratorUseCase) {
	fmt.Printf("%-10s %-10s %-9s %-20s %-6s %-12s %s\n", "AGENTE", "INSTALADO", "ACTIVO", "TIERS", "PRIO", "HOY/LÍMITE", "COOLDOWN")
	for _, s := range orchestrator.Status() {
		limit := "∞"
		if s.Profile.DailyLimit > 0 {
			limit = fmt.Sprint(s.Profile.DailyLimit)
		}
		cooldown := "-"
		if time.Now().Before(s.CooldownUntil) {
			cooldown = "hasta " + s.CooldownUntil.Format("15:04")
		}
		tiers := make([]string, len(s.Profile.Tiers))
		for i, t := range s.Profile.Tiers {
			tiers[i] = string(t)
		}
		fmt.Printf("%-10s %-10v %-9v %-20s %-6d %-12s %s\n",
			s.Profile.Name, s.Installed, s.Profile.Enabled, strings.Join(tiers, ","),
			s.Profile.Priority, fmt.Sprintf("%d/%s", s.RunsToday, limit), cooldown)
	}
}

func serve(chatProxyUseCase *usecases.ChatProxyUseCase, orchestrator *usecases.OrchestratorUseCase) {
	port := 3500

	// ==========================================
	// 2. CAPA DE PRESENTACIÓN (API / Enrutador)
	// ==========================================

	http.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		// Validamos que sea un POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 1. Parseamos el JSON que viene de Claude Code/Usuario al Request Model
		var req request.ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// 2. Ejecutamos nuestro núcleo de negocio limpio
		response := chatProxyUseCase.Execute(req)

		// 3. Devolvemos la respuesta formateada al cliente
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// Delegación de tareas al orquestador multi-agente
	http.HandleFunc("/v1/orchestrate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req request.OrchestrateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Prompt == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		result, err := orchestrator.Execute(r.Context(), req.Prompt, req.WorkDir, req.Agent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// ==========================================
	// 3. ARRANQUE DEL SERVIDOR
	// ==========================================
	fmt.Println("🚀 TAMI Core iniciando con Clean Architecture...")
	fmt.Printf("🌐 Gateway Multi-Agente escuchando en: http://localhost:%d\n", port)

	addr := fmt.Sprintf(":%d", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("❌ Error fatal: %v", err)
	}
}
