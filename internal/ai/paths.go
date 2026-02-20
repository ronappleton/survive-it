package ai

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	ModelURL      = "https://example.com/ai-pack-v1.gguf"
	modelFileName = "ai-pack-v1.gguf"
)

func appSupportDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if home == "" {
		return "", errors.New("home directory not found")
	}
	return filepath.Join(home, "Library", "Application Support", "SurviveIt"), nil
}

func ConfigPath() (string, error) {
	dir, err := appSupportDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func modelsDirPath() (string, error) {
	dir, err := appSupportDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "models"), nil
}

func ModelPath() (string, error) {
	dir, err := modelsDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, modelFileName), nil
}
