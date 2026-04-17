package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Ars-Ludus/providertron/models"
)

func main() {
	cfg := parseFlags()
	setupLogger(cfg.LogLevel)

	if cfg.Provider == "" || cfg.APIKey == "" || cfg.Command == "" {
		fmt.Fprintln(os.Stderr, "error: -provider, -api-key, and -command are required")
		flag.Usage()
		os.Exit(2)
	}

	p, err := buildProvider(cfg)
	if err != nil {
		slog.Default().Error("failed to build provider", "err", err)
		os.Exit(1)
	}

	store := &models.Store{Path: cfg.ModelsPath}
	ctx := context.Background()

	var cmdErr error
	switch cfg.Command {
	case "generate":
		cmdErr = runGenerate(ctx, p, cfg)
	case "stream":
		cmdErr = runStream(ctx, p, cfg)
	case "embed":
		cmdErr = runEmbed(ctx, p, cfg)
	case "batch":
		cmdErr = runBatch(ctx, p, cfg)
	case "voice":
		cmdErr = runVoice(ctx, p, cfg)
	case "discover":
		cmdErr = runDiscover(ctx, p, store)
	case "dispatch":
		cmdErr = runDispatch(ctx, p, cfg)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n", cfg.Command)
		flag.Usage()
		os.Exit(2)
	}

	if cmdErr != nil {
		slog.Default().Error("command failed", "command", cfg.Command, "err", cmdErr)
		os.Exit(1)
	}
}

func setupLogger(level string) {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: l,
	})))
}
