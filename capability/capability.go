package capability

import "context"

type Generator interface {
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
}

type Streamer interface {
	Stream(ctx context.Context, req StreamRequest) (<-chan StreamChunk, error)
}

type Embedder interface {
	Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error)
}

type Batcher interface {
	Batch(ctx context.Context, req BatchRequest) (BatchResponse, error)
}

type Voicer interface {
	Voice(ctx context.Context, req VoiceRequest) (VoiceResponse, error)
}

type Discoverer interface {
	FetchModels(ctx context.Context) ([]ModelInfo, error)
}
