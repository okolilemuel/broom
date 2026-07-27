# 🧹 broom

> Sweep your dev machine clean.

A guided terminal UI for reclaiming disk space eaten by the side effects of being a developer — `node_modules`, Python venvs, Rust build artifacts, LLM model files, Docker images, build caches, and more.

```
╭────────────────────────────────────────────────────────────╮
│                                                            │
│   🧹 broom                                                 │
│  Sweep your dev machine clean                              │
│                                                            │
│  Scans for:                                                │
│    • node_modules & Python venvs                           │
│    • Rust target dirs & Cargo cache                        │
│    • Build output (.next, dist, build…)                    │
│    • LLM model files (HuggingFace, LM Studio, Ollama)      │
│    • Package caches (npm, uv, pnpm, Homebrew…)             │
│    • Docker / Podman images & volumes                      │
│    • Xcode DerivedData & simulators                        │
│    • App caches (Puppeteer, Playwright…)                   │
│                                                            │
│  [ press enter to start scan ]                             │
│    q to quit                                               │
│                                                            │
╰────────────────────────────────────────────────────────────╯
```

---

## The story

One morning I opened my Mac and saw this:

```
Filesystem    Size    Used   Avail  Capacity
/dev/disk3s5  926Gi  711Gi  150Gi     83%
```

**711 GB used. 150 GB free. On a 926 GB drive.**

I'd been building things for months — a Solana project, a few AI agents, some fintech APIs — the usual. Docker containers, Python venvs, `node_modules` across dozens of projects, LLM models I downloaded once and forgot about. The machine was fine. I just never cleaned up after myself.

So I did it manually. Ran `du -sh` on everything. Found 44 GB of Docker build cache I didn't need. Found 19 GB in a single `buildx_buildkit_default_state` volume. Found 15 GB of HuggingFace models from experiments I'd long abandoned — Bark TTS, MusicGen, NLLB translation. Found `node_modules` in projects last touched in 2022. Python venvs from 2024. Rust `target/` directories from a Solana validator I hadn't run in months.

By the end I'd freed **270 GB**. In one session.

The whole time I kept thinking: *this should be a tool.* Not a script that blindly deletes things, but something guided — something that shows you what it found, how big it is, how old it is, and lets you decide. Something that runs in under a minute and asks the right questions.

That's broom. I built it to scratch my own itch, and then I put it out there because every developer I know has the same problem. We install things, experiment, move on, and forget. The disk quietly fills up. Eventually something breaks or you get that low-storage warning and spend an afternoon doing archaeology.

broom does the archaeology for you.

---

## Installation

### Homebrew (recommended)

```bash
brew tap okolilemuel/tap
brew trust okolilemuel/tap
brew install broom
```

### One-line install script

```bash
curl -fsSL https://raw.githubusercontent.com/okolilemuel/broom/main/install.sh | bash
```

### Build from source

```bash
git clone https://github.com/okolilemuel/broom.git
cd broom
make install   # builds and copies to /usr/local/bin
```

Requires Go 1.22+.

---

## Usage

```bash
broom                      # interactive scan → select → delete
broom --dry-run            # scan and preview without deleting anything
broom --older-than 6m      # auto-select items older than 6 months
broom --older-than 1y      # or 1 year, 90d, 180d…
broom --version
broom --help
```

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `↑` / `↓` or `j` / `k` | Navigate list |
| `space` | Toggle item selection |
| `tab` | Toggle entire category |
| `a` | Select all visible items |
| `A` | Deselect all visible items |
| `/` | Fuzzy filter |
| `esc` | Exit filter mode |
| `enter` | Proceed to confirm |
| `q` | Quit |

---

## What broom scans

### Project artifacts

| Type | Finds |
|------|-------|
| **Node modules** | `node_modules/` directories in project trees, sorted by size |
| **Python venvs** | `.venv/`, `venv/`, `env/` — confirmed as real venvs before listing |
| **Rust targets** | `target/` directories next to a `Cargo.toml` |
| **Build output** | `.next/`, `dist/`, `build/`, `.turbo/`, `.parcel-cache/`, `out/` |

All project artifacts show **last git commit date** so you know exactly how stale they are. Items with no git history are flagged as `no git`.

### Global caches

| Cache | Location |
|-------|----------|
| npm | `~/.npm` |
| uv | `~/.cache/uv` |
| pnpm store | `~/Library/pnpm/store` |
| Cargo registry | `~/.cargo/registry/src` + `~/.cargo/git` |
| Homebrew | `~/Library/Caches/Homebrew` |
| Puppeteer | `~/.cache/puppeteer` |
| Playwright | `~/Library/Caches/ms-playwright` |
| Gradle | `~/.gradle/caches` |
| Maven | `~/.m2/repository` |
| Solana | `~/.cache/solana` |

### LLM models

| Source | What's found |
|--------|-------------|
| HuggingFace | Each model under `~/.cache/huggingface/hub/` listed individually |
| LM Studio | GGUF files under `~/.cache/lm-studio/models/` |
| Ollama | `~/.ollama/models/` blob store |

### Docker / Podman

Works with both Docker Desktop and Podman (they share a backend on macOS).

- Dangling images (untagged build layers)
- Stopped containers
- Unused volumes
- BuildKit / buildx cache
- Unused named images older than 30 days

### Xcode (macOS)

- DerivedData
- Build Archives
- iOS Device Support symbols
- Unavailable simulators (`xcrun simctl delete unavailable`)

---

## Ignore list

Create `~/.config/broom/ignore` to permanently skip paths:

```
# broom ignore file
# one path prefix per line, ~ is expanded

~/Projects/active-client
~/Projects/company-monorepo
~/work/do-not-touch
```

Any directory whose path starts with an entry in this file will be skipped during scanning.

---

## Dry run

Always safe to run `--dry-run` first. It scans everything, shows you exactly what it would delete, walks you through the selection flow, and prints a summary — without touching a single file.

```bash
broom --dry-run
```

The UI shows a `[DRY RUN]` badge throughout and the done screen says **"Would free X GB"** instead of **"Freed X GB"**.

---

## Auto-selection

The `--older-than` flag pre-selects items based on age so you don't have to manually pick through everything:

```bash
broom --older-than 6m    # pre-select anything last committed >6 months ago
broom --older-than 1y    # pre-select anything >1 year old
```

Items with **no git history** are always auto-selected when this flag is set — they're almost always throwaway experiments or demo projects.

You can still deselect anything before confirming.

---

## How it works

broom scans concurrently — all scanners run in parallel goroutines and stream results into the UI in real time. You see items appear as they're found rather than waiting for a full scan to complete.

Deduplication is handled via a shared `sync.Map` keyed on the real filesystem path (`filepath.EvalSymlinks`), so even if you have symlinks or multiple search roots that overlap, the same directory is never listed twice.

Deletion respects each item's `CleanFunc` — Docker items run `docker system prune`, Xcode simulators run `xcrun simctl delete unavailable`, everything else is a standard `os.RemoveAll`. The `--dry-run` flag skips all of these.

---

## Building from source

```bash
git clone https://github.com/okolilemuel/broom.git
cd broom

make build      # ./broom
make install    # /usr/local/bin/broom
make release    # all 4 platform binaries
make clean
```

Cross-compilation targets:

| Binary | Platform |
|--------|----------|
| `broom-darwin-arm64` | macOS Apple Silicon |
| `broom-darwin-amd64` | macOS Intel |
| `broom-linux-arm64` | Linux ARM64 |
| `broom-linux-amd64` | Linux x86-64 |

---

## Releasing

Tag a version and the GitHub Actions workflow handles the rest:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The workflow:
1. Builds all 4 platform binaries
2. Creates a GitHub Release with binaries attached
3. Computes SHA256 checksums and automatically updates the Homebrew tap formula

`brew upgrade broom` picks it up for all users within minutes.

---

## Platform support

| Platform | Status |
|----------|--------|
| macOS (Apple Silicon) | ✅ Full support |
| macOS (Intel) | ✅ Full support |
| Linux (arm64) | ✅ Supported (Xcode scanner is a no-op) |
| Linux (amd64) | ✅ Supported (Xcode scanner is a no-op) |
| Windows | 🔜 Planned |

---

## Contributing

Pull requests are welcome. A few areas where contributions would be especially useful:

- **Linux-specific paths** — Linux has different cache locations for tools like pnpm, Gradle, pip. If you use Linux and know the right paths, open a PR adding them to `internal/platform/linux.go`.
- **Windows support** — `internal/platform/windows.go` doesn't exist yet. The platform interface is designed to make this straightforward.
- **New scanners** — Maven, Pip download cache, Android SDK, Go module cache, Terraform providers, etc. Each scanner is a small self-contained file in `internal/scan/`.
- **Homebrew tap automation** — The formula auto-update workflow is a single shell script in `release.yml`. Improvements welcome.

To add a new scanner, implement the `scan.Scanner` interface:

```go
type Scanner interface {
    Name() string
    Scan(ctx context.Context, roots []platform.Root, sc ScanContext, out chan<- Item)
}
```

Then register it in `internal/ui/model.go`'s `startScanning()`.

---

## License

MIT — see [LICENSE](LICENSE).

---

*Built by [Lemuel Okoli](https://github.com/okolilemuel). If broom saved you disk space, consider giving it a ⭐.*
