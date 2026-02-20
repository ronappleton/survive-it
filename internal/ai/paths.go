package ai

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type ModelPack struct {
	ID       string
	Name     string
	FileName string
	URL      string
	SizeGB   float64
	Speed    string
	Quality  string
	BestFor  string
	Summary  string
}

var modelPacks = []ModelPack{
	{
		ID:       "qwen2_5_0_5b_q4km",
		Name:     "Qwen2.5 0.5B Instruct (Q4_K_M)",
		FileName: "Qwen2.5-0.5B-Instruct-Q4_K_M.gguf",
		URL:      "https://huggingface.co/bartowski/Qwen2.5-0.5B-Instruct-GGUF/resolve/main/Qwen2.5-0.5B-Instruct-Q4_K_M.gguf?download=true",
		SizeGB:   0.37,
		Speed:    "Fast",
		Quality:  "Good",
		BestFor:  "Low-memory systems",
		Summary:  "Smallest and quickest to run. Best when you want responsive AI with minimal resource use.",
	},
	{
		ID:       "llama3_2_1b_q4km",
		Name:     "Llama 3.2 1B Instruct (Q4_K_M)",
		FileName: "Llama-3.2-1B-Instruct-Q4_K_M.gguf",
		URL:      "https://huggingface.co/bartowski/Llama-3.2-1B-Instruct-GGUF/resolve/main/Llama-3.2-1B-Instruct-Q4_K_M.gguf?download=true",
		SizeGB:   0.75,
		Speed:    "Medium",
		Quality:  "Better",
		BestFor:  "Everyday play",
		Summary:  "Balanced option. Better answers than tiny models while still running well on most machines.",
	},
	{
		ID:       "phi3_mini_4k_q4km",
		Name:     "Phi-3 Mini 4K Instruct (Q4_K_M)",
		FileName: "Phi-3-mini-4k-instruct-Q4_K_M.gguf",
		URL:      "https://huggingface.co/bartowski/Phi-3-mini-4k-instruct-GGUF/resolve/main/Phi-3-mini-4k-instruct-Q4_K_M.gguf?download=true",
		SizeGB:   2.23,
		Speed:    "Slower",
		Quality:  "Best",
		BestFor:  "Higher-quality replies",
		Summary:  "Largest model here. Best quality, but takes more disk and may run slower.",
	},
}

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

func AvailableModelPacks() []ModelPack {
	out := make([]ModelPack, len(modelPacks))
	copy(out, modelPacks)
	return out
}

func DefaultModelID() string {
	if len(modelPacks) == 0 {
		return ""
	}
	return modelPacks[0].ID
}

func NormalizeModelID(id string) string {
	id = strings.TrimSpace(strings.ToLower(id))
	if id == "" {
		return DefaultModelID()
	}
	if _, ok := ModelPackByID(id); ok {
		return id
	}
	return DefaultModelID()
}

func ModelPackByID(id string) (ModelPack, bool) {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, pack := range modelPacks {
		if pack.ID == id {
			return pack, true
		}
	}
	return ModelPack{}, false
}

func ModelPathForID(modelID string) (string, error) {
	pack, ok := ModelPackByID(NormalizeModelID(modelID))
	if !ok {
		return "", errors.New("model pack not found")
	}
	dir, err := modelsDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, pack.FileName), nil
}
