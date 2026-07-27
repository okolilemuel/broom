//go:build darwin

package scan

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"broom/internal/platform"
)

type XcodeScanner struct{}

func (s *XcodeScanner) Name() string { return "xcode simulators" }

func (s *XcodeScanner) Scan(ctx context.Context, roots []platform.Root, sc ScanContext, out chan<- Item) {
	// Run xcrun simctl list devices --json
	raw, err := exec.CommandContext(ctx, "xcrun", "simctl", "list", "devices", "--json").Output()
	if err != nil {
		return
	}

	var result struct {
		Devices map[string][]struct {
			UDID        string `json:"udid"`
			IsAvailable bool   `json:"isAvailable"`
			State       string `json:"state"`
		} `json:"devices"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return
	}

	home, _ := os.UserHomeDir()
	devicesBase := filepath.Join(home, "Library", "Developer", "CoreSimulator", "Devices")

	var totalSize int64
	for _, devices := range result.Devices {
		for _, dev := range devices {
			if dev.IsAvailable {
				continue
			}
			devPath := filepath.Join(devicesBase, dev.UDID)
			if !PathExists(devPath) {
				continue
			}
			totalSize += DirSizeBytes(devPath)
		}
	}

	if totalSize == 0 {
		return
	}

	// Deduplicate by a stable key
	key := "xcode:unavailable-simulators"
	if _, loaded := sc.Seen.LoadOrStore(key, true); loaded {
		return
	}

	out <- Item{
		Path:        devicesBase,
		DisplayName: "Xcode unavailable simulators",
		Category:    CategoryXcode,
		SizeBytes:   totalSize,
		HasGit:      false,
		Description: "unavailable simulator runtimes",
		CleanFunc: func() error {
			return exec.Command("xcrun", "simctl", "delete", "unavailable").Run()
		},
	}
}
