//go:build darwin

package platform

import (
	"os"
	"path/filepath"
	"syscall"
)

type darwin struct{ home string }

func New() Platform {
	home, _ := os.UserHomeDir()
	return &darwin{home: home}
}

func (d *darwin) Name() string { return "macOS" }

func (d *darwin) SearchRoots() []Root {
	roots := []Root{}
	// well-known project directory names
	projectDirs := []string{
		"Projects", "projects", "Development", "Dev", "dev",
		"Code", "code", "Work", "work", "src", "Src",
		"GitHub", "GitLab", "repos", "Repos",
	}
	for _, name := range projectDirs {
		p := filepath.Join(d.home, name)
		if stat, err := os.Stat(p); err == nil && stat.IsDir() {
			roots = append(roots, Root{Path: p, MaxDepth: 6})
		}
	}
	// also scan home directly at shallow depth to catch top-level project folders
	roots = append(roots, Root{Path: d.home, MaxDepth: 2})
	return roots
}

func (d *darwin) GlobalCaches() []CacheEntry {
	h := d.home
	join := filepath.Join
	return []CacheEntry{
		// package managers
		{Name: "npm cache", Path: join(h, ".npm"), Category: "Package Caches"},
		{Name: "uv cache", Path: join(h, ".cache", "uv"), Category: "Package Caches"},
		{Name: "pnpm store", Path: join(h, "Library", "pnpm", "store"), Category: "Package Caches"},
		{Name: "Cargo registry", Path: join(h, ".cargo", "registry", "src"), Category: "Package Caches"},
		{Name: "Cargo git cache", Path: join(h, ".cargo", "git"), Category: "Package Caches"},
		{Name: "Homebrew cache", Path: join(h, "Library", "Caches", "Homebrew"), Category: "Package Caches"},
		// LLM models
		{Name: "HuggingFace cache", Path: join(h, ".cache", "huggingface", "hub"), Category: "LLM Models"},
		{Name: "LM Studio models", Path: join(h, ".cache", "lm-studio", "models"), Category: "LLM Models"},
		{Name: "Ollama models", Path: join(h, ".ollama", "models"), Category: "LLM Models"},
		// browser automation
		{Name: "Puppeteer browsers", Path: join(h, ".cache", "puppeteer"), Category: "App Caches"},
		{Name: "Playwright browsers", Path: join(h, "Library", "Caches", "ms-playwright"), Category: "App Caches"},
		// Xcode
		{Name: "Xcode DerivedData", Path: join(h, "Library", "Developer", "Xcode", "DerivedData"), Category: "Xcode"},
		{Name: "Xcode Archives", Path: join(h, "Library", "Developer", "Xcode", "Archives"), Category: "Xcode"},
		{Name: "iOS Device Support", Path: join(h, "Library", "Developer", "Xcode", "iOS DeviceSupport"), Category: "Xcode"},
		// iOS simulator
		{Name: "Simulator data", Path: join(h, "Library", "Developer", "CoreSimulator", "Devices"), Category: "Xcode"},
		// misc dev caches
		{Name: "Solana cache", Path: join(h, ".cache", "solana"), Category: "App Caches"},
		{Name: "Codex sessions", Path: join(h, ".codex", "sessions"), Category: "App Caches"},
		{Name: "act cache", Path: join(h, ".cache", "act"), Category: "App Caches"},
		{Name: "Gradle caches", Path: join(h, ".gradle", "caches"), Category: "App Caches"},
		{Name: "Maven repository", Path: join(h, ".m2", "repository"), Category: "App Caches"},
	}
}

func (d *darwin) DiskFreeBytes(path string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}
