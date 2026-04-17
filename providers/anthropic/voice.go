package anthropic

import (
	"context"
	"fmt"

	"github.com/arsmorientem/providertron/capability"
)

func (a *Anthropic) Voice(_ context.Context, _ capability.VoiceRequest) (capability.VoiceResponse, error) {
	return capability.VoiceResponse{}, fmt.Errorf(
		"anthropic.Voice: Anthropic does not provide a voice API; "+
			"use OpenAI for TTS/STT or Gemini for TTS: %w",
		capability.ErrCapabilityUnavailable,
	)
}
