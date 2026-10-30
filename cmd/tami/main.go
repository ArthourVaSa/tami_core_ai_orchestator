package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	// Ajusta estos imports según el nombre real de tu módulo en go.mod (ej. "tami-core/...")
	"tami-core/internal/domain/models/request"
	"tami-core/internal/infrastructure/repositories"
	"tami-core/internal/infrastructure/services"
	usecases "tami-core/internal/use_cases"
)

func main() {
	port := 3500

	// ==========================================
	// 1. INYECCIÓN DE DEPENDENCIAS (Wiring)
	// ==========================================

	// a. Instanciamos la Infraestructura
	llmService := services.NewOllamaService("http://localhost:11434", "llama3.1:latest") // Cambia "llama3" por el modelo que uses
	memoryRepo := repositories.NewChromaRepo("http://localhost:8000", "tami_memory")

	// b. Instanciamos el Caso de Uso inyectándole la infraestructura
	// (Aquí Go verifica en tiempo de compilación que cumplan las interfaces)
	chatProxyUseCase := usecases.NewChatProxyUseCase(memoryRepo, llmService)

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
