package gemini

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arsmorientem/providertron/capability"
)

const (
	defaultVoiceModel = "gemini-2.5-flash-preview-tts"
)

type ttsRequestBody struct {
	Contents         []geminiContent `json:"contents"`
	GenerationConfig ttsGenConfig    `json:"generationConfig"`
}

type ttsGenConfig struct {
	ResponseModalities []string    `json:"responseModalities"`
	SpeechConfig       speechConfig `json:"speechConfig"`
}

type speechConfig struct {
	VoiceConfig voiceConfig `json:"voiceConfig"`
}

type voiceConfig struct {
	PrebuiltVoiceConfig prebuiltVoiceConfig `json:"prebuiltVoiceConfig"`
}

type prebuiltVoiceConfig struct {
	VoiceName string `json:"voiceName"`
}

type ttsResponseBody struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				InlineData *struct {
					MimeType string `json:"mimeType"`
					Data     []byte `json:"data"`
				} `json:"inlineData,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (g *Gemini) Voice(ctx context.Context, req capability.VoiceRequest) (capability.VoiceResponse, error) {
	switch req.Direction {
	case "tts":
		return g.tts(ctx, req)
	case "stt":
		return capability.VoiceResponse{}, fmt.Errorf(
			"gemini.Voice: speech-to-text is not available via the REST API; "+
				"use the Live API (WebSocket) for real-time STT: %w",
			capability.ErrCapabilityUnavailable,
		)
	default:
		return capability.VoiceResponse{}, fmt.Errorf(
			"gemini.Voice: unknown direction %q, expected \"tts\" or \"stt\": %w",
			req.Direction, capability.ErrCapabilityUnavailable,
		)
	}
}

func (g *Gemini) tts(ctx context.Context, req capability.VoiceRequest) (capability.VoiceResponse, error) {
	model := req.Model
	if model == "" {
		model = g.cfg.VoiceModel
	}
	if model == "" {
		model = defaultVoiceModel
	}

	voiceName := req.Voice
	if voiceName == "" {
		voiceName = "Kore"
	}

	url := g.buildURL("/v1beta/models/" + model + ":generateContent")

	body := ttsRequestBody{
		Contents: []geminiContent{
			{Role: "user", Parts: []geminiPart{{Text: req.Input}}},
		},
		GenerationConfig: ttsGenConfig{
			ResponseModalities: []string{"AUDIO"},
			SpeechConfig: speechConfig{
				VoiceConfig: voiceConfig{
					PrebuiltVoiceConfig: prebuiltVoiceConfig{VoiceName: voiceName},
				},
			},
		},
	}

	slog.Default().With("pkg", "gemini").Info("tts", "model", model, "voice", voiceName)

	var raw ttsResponseBody
	if err := g.doRequest(ctx, url, body, &raw); err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("gemini.Voice tts: %w", err)
	}

	if len(raw.Candidates) == 0 || len(raw.Candidates[0].Content.Parts) == 0 {
		return capability.VoiceResponse{}, fmt.Errorf("gemini.Voice tts: empty response")
	}

	part := raw.Candidates[0].Content.Parts[0]
	if part.InlineData == nil {
		return capability.VoiceResponse{}, fmt.Errorf("gemini.Voice tts: no audio data in response")
	}

	return capability.VoiceResponse{
		Direction: "tts",
		Audio:     part.InlineData.Data,
		Model:     model,
	}, nil
}
