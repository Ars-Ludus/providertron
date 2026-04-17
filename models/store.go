package models

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Ars-Ludus/providertron/capability"
)

const schemaVersion = "1"

type Store struct {
	Path string
}

func (s *Store) Load() (capability.ModelsFile, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Default().With("pkg", "models").Debug("models file not found, starting empty", "path", s.Path)
			return capability.ModelsFile{
				Version: schemaVersion,
				Models:  make(map[string]capability.ModelInfo),
			}, nil
		}
		return capability.ModelsFile{}, fmt.Errorf("models.Store.Load: %w", err)
	}

	var f capability.ModelsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return capability.ModelsFile{}, fmt.Errorf("models.Store.Load: decode: %w", err)
	}
	if f.Models == nil {
		f.Models = make(map[string]capability.ModelInfo)
	}

	slog.Default().With("pkg", "models").Debug("loaded models file", "path", s.Path, "count", len(f.Models))
	return f, nil
}

func (s *Store) Save(f capability.ModelsFile) error {
	f.Version = schemaVersion
	f.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("models.Store.Save: marshal: %w", err)
	}

	dir := filepath.Dir(s.Path)
	tmp, err := os.CreateTemp(dir, ".models-*.json.tmp")
	if err != nil {
		return fmt.Errorf("models.Store.Save: create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("models.Store.Save: write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("models.Store.Save: close temp: %w", err)
	}

	if err := os.Rename(tmpName, s.Path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("models.Store.Save: rename: %w", err)
	}

	slog.Default().With("pkg", "models").Info("saved models file", "path", s.Path, "count", len(f.Models))
	return nil
}

func Merge(existing, incoming capability.ModelsFile) capability.ModelsFile {
	if existing.Models == nil {
		existing.Models = make(map[string]capability.ModelInfo)
	}
	for k, v := range incoming.Models {
		existing.Models[k] = v
	}
	return existing
}
