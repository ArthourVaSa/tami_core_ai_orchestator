package request

type OrchestrateRequest struct {
	Prompt  string `json:"prompt"`
	WorkDir string `json:"work_dir,omitempty"`
	Agent   string `json:"agent,omitempty"` // opcional: fuerza un agente concreto
}
