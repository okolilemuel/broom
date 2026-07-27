package scan

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"broom/internal/platform"
)

const minModelSize = 100 * 1024 * 1024 // 100 MB

type LLMScanner struct{}

func (s *LLMScanner) Name() string { return "llm models" }

func (s *LLMScanner) Scan(ctx context.Context, roots []platform.Root, out chan<- Item) {
	home := homeDir()

	// HuggingFace hub: scan each model separately
	hfHub := filepath.Join(home, ".cache", "huggingface", "hub")
	if PathExists(hfHub) {
		entries, err := os.ReadDir(hfHub)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				name := e.Name()
				if !strings.HasPrefix(name, "models--") {
					continue
				}
				path := filepath.Join(hfHub, name)
				size := DirSizeBytes(path)
				if size < minModelSize {
					continue
				}
				// convert models--org--name → org/name
				displayName := strings.TrimPrefix(name, "models--")
				displayName = strings.ReplaceAll(displayName, "--", "/")
				out <- Item{
					Path:        path,
					DisplayName: displayName,
					Category:    CategoryLLMModel,
					SizeBytes:   size,
					HasGit:      false,
					Description: "HuggingFace model",
				}
			}
		}
	}

	// LM Studio: scan each model file
	lmStudio := filepath.Join(home, ".cache", "lm-studio", "models")
	scanGGUFModels(lmStudio, "LM Studio", out)

	// Ollama blobs
	ollamaModels := filepath.Join(home, ".ollama", "models")
	if PathExists(ollamaModels) {
		size := DirSizeBytes(ollamaModels)
		if size >= minModelSize {
			out <- Item{
				Path:        ollamaModels,
				DisplayName: "Ollama models",
				Category:    CategoryLLMModel,
				SizeBytes:   size,
				HasGit:      false,
				Description: "Ollama model blobs",
			}
		}
	}
}

func scanGGUFModels(root, source string, out chan<- Item) {
	if !PathExists(root) {
		return
	}
	// walk up to 3 levels looking for .gguf files
	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		if depth > 3 {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			path := filepath.Join(dir, e.Name())
			if e.IsDir() {
				walk(path, depth+1)
				continue
			}
			if strings.HasSuffix(strings.ToLower(e.Name()), ".gguf") {
				size := DirSizeBytes(path)
				if size < minModelSize {
					continue
				}
				out <- Item{
					Path:        path,
					DisplayName: e.Name(),
					Category:    CategoryLLMModel,
					SizeBytes:   size,
					HasGit:      false,
					Description: source + " model",
				}
			}
		}
	}
	walk(root, 0)
}
