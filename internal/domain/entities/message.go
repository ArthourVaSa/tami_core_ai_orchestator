package entities

// Message es la entidad central compartida entre peticiones y respuestas
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
