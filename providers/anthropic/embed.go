package anthropic

import (
	"context"
	"fmt"

	"github.com/arsmorientem/providertron/capability"
)

func (a *Anthropic) Embed(_ context.Context, _ capability.EmbedRequest) (capability.EmbedResponse, error) {
	return capability.EmbedResponse{}, fmt.Errorf(
		"anthropic.Embed: Anthropic does not provide a text embedding API; "+
			"use OpenAI or Gemini for embeddings: %w",
		capability.ErrCapabilityUnavailable,
	)
}
