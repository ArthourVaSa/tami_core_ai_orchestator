package services

// LLMResponse encapsula la salida bruta del motor
type LLMResponse struct {
	Content   string
	ToolCalls []ToolCallDef
}
