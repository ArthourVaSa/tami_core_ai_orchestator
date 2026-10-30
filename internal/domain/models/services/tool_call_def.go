package services

// ToolCallDef representa la intención del LLM de ejecutar una herramienta
type ToolCallDef struct {
	Name      string
	Arguments string
}
