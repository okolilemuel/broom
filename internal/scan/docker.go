package scan

import (
	"context"
	"os/exec"
	"strconv"
	"strings"

	"broom/internal/platform"
)

type DockerScanner struct{}

func (s *DockerScanner) Name() string { return "docker/podman" }

func (s *DockerScanner) Scan(ctx context.Context, roots []platform.Root, out chan<- Item) {
	for _, cli := range []string{"docker", "podman"} {
		if !cliExists(cli) {
			continue
		}

		// parse `docker/podman system df` once
		reclaimable := systemDF(cli)

		if v, ok := reclaimable["images"]; ok && v > 0 {
			out <- Item{
				Path:        cli + ":images:dangling",
				DisplayName: cli + " dangling images",
				Category:    CategoryDocker,
				SizeBytes:   v,
				HasGit:      false,
				Description: "untagged / unused image layers",
				CleanFunc: func() error {
					return exec.Command(cli, "image", "prune", "--force").Run()
				},
			}
		}
		if v, ok := reclaimable["containers"]; ok && v > 0 {
			out <- Item{
				Path:        cli + ":containers:stopped",
				DisplayName: cli + " stopped containers",
				Category:    CategoryDocker,
				SizeBytes:   v,
				HasGit:      false,
				Description: "exited containers",
				CleanFunc: func() error {
					return exec.Command(cli, "container", "prune", "--force").Run()
				},
			}
		}
		if v, ok := reclaimable["volumes"]; ok && v > 0 {
			out <- Item{
				Path:        cli + ":volumes",
				DisplayName: cli + " unused volumes",
				Category:    CategoryDocker,
				SizeBytes:   v,
				HasGit:      false,
				Description: "volumes not attached to any container",
				CleanFunc: func() error {
					return exec.Command(cli, "volume", "prune", "--force").Run()
				},
			}
		}

		// build cache via buildx du
		if bcSize := buildxCacheSize(cli); bcSize > 0 {
			out <- Item{
				Path:        cli + ":buildcache",
				DisplayName: cli + " build cache",
				Category:    CategoryDocker,
				SizeBytes:   bcSize,
				HasGit:      false,
				Description: "BuildKit layer cache",
				CleanFunc: func() error {
					return exec.Command(cli, "buildx", "prune", "--force").Run()
				},
			}
		}

		break // docker and podman share a backend; scan once
	}
}

// systemDF runs `docker system df` and returns reclaimable bytes per type.
func systemDF(cli string) map[string]int64 {
	out, err := exec.Command(cli, "system", "df").Output()
	if err != nil {
		return nil
	}
	result := map[string]int64{}
	for _, line := range strings.Split(string(out), "\n") {
		lower := strings.ToLower(line)
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		// last column is "Reclaimable" which may look like "19.41GB (65%)"
		reclaimField := fields[len(fields)-1]
		if strings.Contains(reclaimField, "%") && len(fields) >= 2 {
			reclaimField = fields[len(fields)-2]
		}
		bytes := parseDockerSize(reclaimField)
		switch {
		case strings.Contains(lower, "image"):
			result["images"] = bytes
		case strings.Contains(lower, "container"):
			result["containers"] = bytes
		case strings.Contains(lower, "volume"):
			result["volumes"] = bytes
		}
	}
	return result
}

func buildxCacheSize(cli string) int64 {
	out, err := exec.Command(cli, "buildx", "du").Output()
	if err != nil {
		return 0
	}
	// last line is usually "Total: <size> reclaimable: <size>"
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "total") {
			fields := strings.Fields(line)
			for i, f := range fields {
				if strings.EqualFold(f, "total:") && i+1 < len(fields) {
					return parseDockerSize(fields[i+1])
				}
			}
		}
	}
	return 0
}

func cliExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// parseDockerSize converts Docker human-readable sizes (e.g. "1.5GB", "942MB") to bytes.
func parseDockerSize(s string) int64 {
	s = strings.TrimSpace(s)
	// strip trailing parenthetical
	if idx := strings.Index(s, "("); idx != -1 {
		s = strings.TrimSpace(s[:idx])
	}
	if s == "" || s == "0B" {
		return 0
	}
	units := []struct {
		suffix string
		mult   int64
	}{
		{"TiB", 1 << 40}, {"GiB", 1 << 30}, {"MiB", 1 << 20}, {"KiB", 1 << 10},
		{"TB", 1e12}, {"GB", 1e9}, {"MB", 1e6}, {"KB", 1e3}, {"B", 1},
	}
	upper := strings.ToUpper(s)
	for _, u := range units {
		if strings.HasSuffix(upper, strings.ToUpper(u.suffix)) {
			numStr := s[:len(s)-len(u.suffix)]
			f, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
			if err != nil {
				return 0
			}
			return int64(f * float64(u.mult))
		}
	}
	// plain bytes
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
