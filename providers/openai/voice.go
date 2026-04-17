package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/Ars-Ludus/providertron/capability"
)

type ttsRequest struct {
	Model  string `json:"model"`
	Input  string `json:"input"`
	Voice  string `json:"voice"`
	Format string `json:"response_format,omitempty"`
}

func (o *OpenAI) Voice(ctx context.Context, req capability.VoiceRequest) (capability.VoiceResponse, error) {
	switch req.Direction {
	case "tts":
		return o.tts(ctx, req)
	case "stt":
		return o.stt(ctx, req)
	default:
		return capability.VoiceResponse{}, fmt.Errorf(
			"openai.Voice: unknown direction %q, expected \"tts\" or \"stt\": %w",
			req.Direction, capability.ErrCapabilityUnavailable,
		)
	}
}

func (o *OpenAI) tts(ctx context.Context, req capability.VoiceRequest) (capability.VoiceResponse, error) {
	model := req.Model
	if model == "" {
		model = o.cfg.VoiceModel
	}
	voice := req.Voice
	if voice == "" {
		voice = "alloy"
	}
	format := req.Format
	if format == "" {
		format = "mp3"
	}

	slog.Default().With("pkg", "openai").Info("tts", "model", model, "voice", voice)

	body := ttsRequest{Model: model, Input: req.Input, Voice: voice, Format: format}

	b, err := json.Marshal(body)
	if err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice tts: marshal: %w", err)
	}

	url := o.cfg.BaseURL + "/v1/audio/speech"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice tts: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.cfg.APIKey)

	client := &http.Client{Timeout: time.Duration(o.cfg.HTTPTimeout*3) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice tts: http: %w", err)
	}
	defer resp.Body.Close()

	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice tts: read: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice tts: api error %d: %s", resp.StatusCode, string(audio))
	}

	return capability.VoiceResponse{
		Direction: "tts",
		Audio:     audio,
		Model:     model,
	}, nil
}

func (o *OpenAI) stt(ctx context.Context, req capability.VoiceRequest) (capability.VoiceResponse, error) {
	if len(req.Audio) == 0 {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice stt: audio bytes required")
	}

	model := req.Model
	if model == "" {
		model = o.cfg.STTModel
	}

	slog.Default().With("pkg", "openai").Info("stt", "model", model, "bytes", len(req.Audio))

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	fw, err := mw.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice stt: create form file: %w", err)
	}
	if _, err := fw.Write(req.Audio); err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice stt: write audio: %w", err)
	}
	if err := mw.WriteField("model", model); err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice stt: write model field: %w", err)
	}
	mw.Close()

	url := o.cfg.BaseURL + "/v1/audio/transcriptions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice stt: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+o.cfg.APIKey)

	client := &http.Client{Timeout: time.Duration(o.cfg.HTTPTimeout*3) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice stt: http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice stt: read: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice stt: api error %d: %s", resp.StatusCode, string(raw))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("openai.Voice stt: decode: %w", err)
	}

	return capability.VoiceResponse{
		Direction: "stt",
		Text:      result.Text,
		Model:     model,
	}, nil
}
