package capability

type UsageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type GenerateResponse struct {
	ID      	string    `json:"id"`
	Content 	string    `json:"content"`
	Model   	string    `json:"model"`
	Usage   	UsageInfo `json:"usage"`
}

type StreamChunk struct {
	Delta 		string `json:"delta"`
	Done  		bool   `json:"done"`
	Err   		error  `json:"-"`
}

type EmbedResponse struct {
	Vector 		[]float64 `json:"vector"`
	Model  		string    `json:"model"`
}

type BatchResultItem struct {
	ID     		string `json:"id"`
	Result 		any    `json:"result"`
	Err    		string `json:"error,omitempty"`
}

type BatchResponse struct {
	Items 		[]BatchResultItem `json:"items"`
	Model 		string            `json:"model"`
}

type VoiceResponse struct {
	Direction 	string `json:"direction"`
	Audio     	[]byte `json:"audio,omitempty"`
	Text      	string `json:"text,omitempty"`
	Model     	string `json:"model"`
}
