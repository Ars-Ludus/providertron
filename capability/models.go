package capability

type ModelInfo struct {
	ID           string         `json:"id"`
	Provider     string         `json:"provider"`
	DisplayName  string         `json:"display_name,omitempty"`
	Capabilities []CapabilityType `json:"capabilities"`
	InputTokens  int            `json:"input_tokens,omitempty"`
	OutputTokens int            `json:"output_tokens,omitempty"`
	Deprecated   bool           `json:"deprecated,omitempty"`
}

type ModelsFile struct {
	Version   string               `json:"version"`
	UpdatedAt string               `json:"updated_at"`
	Models    map[string]ModelInfo `json:"models"`
}
