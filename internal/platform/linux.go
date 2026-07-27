//go:build linux

package platform

import (
	"os"
	"path/filepath"
	"syscall"
)

type linux struct{ home string }

func New() Platform {
	home, _ := os.UserHomeDir()
	return &linux{home: home}
}

func (l *linux) Name() string { return "Linux" }

func (l *linux) SearchRoots() []Root {
	roots := []Root{}
	projectDirs := []string{
		"Projects", "projects", "Development", "Dev", "dev",
		"Code", "code", "Work", "work", "src", "Src",
		"GitHub", "GitLab", "repos",
	}
	for _, name := range projectDirs {
		p := filepath.Join(l.home, name)
		if stat, err := os.Stat(p); err == nil && stat.IsDir() {
			roots = append(roots, Root{Path: p, MaxDepth: 6})
		}
	}
	roots = append(roots, Root{Path: l.home, MaxDepth: 2})
	return roots
}

func (l *linux) GlobalCaches() []CacheEntry {
	h := l.home
	join := filepath.Join
	return []CacheEntry{
		{Name: "npm cache", Path: join(h, ".npm"), Category: "Package Caches"},
		{Name: "uv cache", Path: join(h, ".cache", "uv"), Category: "Package Caches"},
		{Name: "pnpm store", Path: join(h, ".local", "share", "pnpm", "store"), Category: "Package Caches"},
		{Name: "Cargo registry", Path: join(h, ".cargo", "registry", "src"), Category: "Package Caches"},
		{Name: "Cargo git cache", Path: join(h, ".cargo", "git"), Category: "Package Caches"},
		{Name: "HuggingFace cache", Path: join(h, ".cache", "huggingface", "hub"), Category: "LLM Models"},
		{Name: "LM Studio models", Path: join(h, ".cache", "lm-studio", "models"), Category: "LLM Models"},
		{Name: "Ollama models", Path: join(h, ".ollama", "models"), Category: "LLM Models"},
		{Name: "Puppeteer browsers", Path: join(h, ".cache", "puppeteer"), Category: "App Caches"},
		{Name: "Playwright browsers", Path: join(h, ".cache", "ms-playwright"), Category: "App Caches"},
		{Name: "Gradle caches", Path: join(h, ".gradle", "caches"), Category: "App Caches"},
		{Name: "Maven repository", Path: join(h, ".m2", "repository"), Category: "App Caches"},
	}
}

func (l *linux) DiskFreeBytes(path string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}
