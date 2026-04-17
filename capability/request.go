package capability

type BaseRequest struct {
	Type CapabilityType `json:"type"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GenerateRequest struct {
	BaseRequest
	Messages    []Message         `json:"messages"`
	Model       string            `json:"model,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
	TopK        int               `json:"top_k,omitempty"`
	TopP        float64           `json:"top_p,omitempty"`
	Stop        []string          `json:"stop,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type StreamRequest struct {
	GenerateRequest
}

type EmbedRequest struct {
	BaseRequest
	Input    string `json:"input"`
	Model    string `json:"model,omitempty"`
	TaskType string `json:"task_type,omitempty"`
}

type BatchItem struct {
	ID    string `json:"id"`
	Input string `json:"input"`
}

type BatchRequest struct {
	BaseRequest
	Items   []BatchItem    `json:"items"`
	Model   string         `json:"model,omitempty"`
	CapType CapabilityType `json:"cap_type"`
}

type VoiceRequest struct {
	BaseRequest
	Direction string `json:"direction"`
	Input     string `json:"input,omitempty"`
	Audio     []byte `json:"audio,omitempty"`
	Model     string `json:"model,omitempty"`
	Voice     string `json:"voice,omitempty"`
	Format    string `json:"format,omitempty"`
}
