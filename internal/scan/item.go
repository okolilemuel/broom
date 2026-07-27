package scan

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
)

type Category string

const (
	CategoryDocker       Category = "Docker / Podman"
	CategoryLLMModel     Category = "LLM Models"
	CategoryPackageCache Category = "Package Caches"
	CategoryNodeModules  Category = "Node Modules"
	CategoryPythonVenv   Category = "Python Venvs"
	CategoryRustArtifact Category = "Rust Artifacts"
	CategoryBuildOutput  Category = "Build Output"
	CategoryXcode        Category = "Xcode"
	CategoryAppCache     Category = "App Caches"
)

// CategoryOrder controls display order in the select screen.
var CategoryOrder = []Category{
	CategoryDocker,
	CategoryLLMModel,
	CategoryPackageCache,
	CategoryNodeModules,
	CategoryPythonVenv,
	CategoryRustArtifact,
	CategoryBuildOutput,
	CategoryXcode,
	CategoryAppCache,
}

// Item represents a single cleanable artifact found on disk.
type Item struct {
	Path        string
	DisplayName string
	Category    Category
	SizeBytes   int64
	LastCommit  time.Time
	HasGit      bool
	Description string
	Selected    bool
	// CleanFunc overrides the default rm -rf behaviour.
	CleanFunc func() error
}

func (i Item) HumanSize() string {
	return humanize.Bytes(uint64(i.SizeBytes))
}

func (i Item) AgeString() string {
	if !i.HasGit {
		return "no git"
	}
	if i.LastCommit.IsZero() {
		return "unknown"
	}
	days := int(time.Since(i.LastCommit).Hours() / 24)
	switch {
	case days < 1:
		return "today"
	case days < 30:
		return fmt.Sprintf("%dd ago", days)
	case days < 365:
		return fmt.Sprintf("%dmo ago", days/30)
	default:
		return fmt.Sprintf("%dy ago", days/365)
	}
}

// FilterValue is used by the fuzzy filter in the select screen.
func (i Item) FilterValue() string {
	return i.DisplayName + " " + string(i.Category) + " " + i.Description
}
