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

> The `brew trust` step is a Homebrew 6.x security requirement for all third-party taps. You only need to run it once.

### One-line install script

```bash
curl -fsSL https://raw.githubusercontent.com/okolilemuel/broom/main/install.sh | bash
```

### Build from source

Requires Go 1.22+.

```bash
git clone https://github.com/okolilemuel/broom.git
cd broom
make build      # builds ./broom
make install    # copies to /usr/local/bin/broom
```

---

## Usage

```bash
broom                      # interactive scan → select → delete
broom --dry-run            # scan and preview without deleting anything
broom --older-than 6m      # auto-select items older than 6 months
broom --older-than 1y      # supports: 6m, 1y, 90d, 180d, or any Go duration
broom --version
broom --help
```

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `↑` / `↓` or `j` / `k` | Navigate list |
| `space` | Toggle item selection |
| `tab` | Toggle all items in current category |
| `a` | Select all visible items |
| `A` | Deselect all visible items |
| `/` | Enter fuzzy filter mode |
| `esc` | Exit filter mode |
| `enter` | Proceed to confirm |
| `q` | Quit |

---

## What broom scans

### Project artifacts

| Type | Finds |
|------|-------|
| **Node modules** | `node_modules/` directories, skips nested ones |
| **Python venvs** | `.venv/`, `venv/`, `env/` — confirmed as real venvs via `bin/python` check |
| **Rust targets** | `target/` directories next to a `Cargo.toml` |
| **Build output** | `.next/`, `dist/`, `build/`, `.turbo/`, `.parcel-cache/`, `out/`, `.output/` |

All project artifacts show the **last git commit date** so you know exactly how stale they are. Items with no git history are flagged as `no git`.

### Package caches

| Cache | macOS path |
|-------|------------|
| npm | `~/.npm` |
| uv | `~/.cache/uv` |
| pnpm store | `~/Library/pnpm/store` |
| Cargo registry | `~/.cargo/registry/src` + `~/.cargo/git` |
| Homebrew | `~/Library/Caches/Homebrew` |
| Gradle | `~/.gradle/caches` |
| Maven | `~/.m2/repository` |

### LLM models

| Source | What's found |
|--------|-------------|
| HuggingFace | Each model under `~/.cache/huggingface/hub/` listed individually by name and size |
| LM Studio | GGUF files under `~/.cache/lm-studio/models/` |
| Ollama | `~/.ollama/models/` blob store |

### App caches

| Cache | Location |
|-------|----------|
| Puppeteer | `~/.cache/puppeteer` |
| Playwright | `~/Library/Caches/ms-playwright` |
| Solana | `~/.cache/solana` |

### Docker / Podman

Works with both Docker Desktop and Podman (they share a backend on macOS).

| Item | Clean action |
|------|-------------|
| Dangling images | `docker image prune --force` |
| Stopped containers | `docker container prune --force` |
| Unused volumes | `docker volume prune --force` |
| BuildKit / buildx cache | `docker buildx prune --force` |
| Unused named images (>30 days old) | `docker rmi <id>` per image |

### Xcode (macOS only)

| Item | Clean action |
|------|-------------|
| DerivedData | `rm -rf` |
| Build Archives | `rm -rf` |
| iOS Device Support | `rm -rf` |
| Unavailable simulators | `xcrun simctl delete unavailable` |

---

## Dry run

Always safe to run `--dry-run` first. broom scans everything, shows you what it found, walks you through the full selection flow, and prints a summary — without touching a single file.

```bash
broom --dry-run
```

The UI shows a `[DRY RUN]` badge throughout and the done screen says **"Would free X GB"** instead of **"Freed X GB"**.

---

## Auto-selection

The `--older-than` flag pre-selects items based on age so you don't have to pick through everything manually:

```bash
broom --older-than 6m    # pre-select anything last committed >6 months ago
broom --older-than 1y    # pre-select anything >1 year old
```

Items with **no git history** are always auto-selected when this flag is set — they're almost always throwaway experiments or demo projects. You can still deselect anything before confirming.

---

## Ignore list

Create `~/.config/broom/ignore` to permanently skip paths during scanning:

```
# broom ignore file — one path prefix per line, ~ is expanded

~/Projects/active-client
~/Projects/company-monorepo
~/work/do-not-touch
```

---

## How it works

broom runs all scanners concurrently in separate goroutines and streams results into the UI in real time — you see items appear as they're found rather than waiting for a full scan to complete.

Deduplication uses a shared `sync.Map` keyed on the resolved filesystem path (`filepath.EvalSymlinks`), so symlinks and overlapping search roots never produce duplicate entries.

Each item carries its own `CleanFunc` — Docker items run targeted prune commands, Xcode simulators run `xcrun simctl delete unavailable`, everything else is `os.RemoveAll`. The `--dry-run` flag skips all clean functions without changing anything else about the flow.

---

## Releasing

Tag a version and the CI handles everything:

```bash
git tag v1.2.0
git push origin v1.2.0
```

The workflow:
1. Builds all 4 platform binaries (darwin/linux × arm64/amd64)
2. Creates a GitHub Release with binaries attached
3. Downloads the binaries, computes SHA256 checksums, and commits an updated formula to the [homebrew-tap](https://github.com/okolilemuel/homebrew-tap) repo automatically

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

- **Linux paths** — Linux has different cache locations for pnpm, pip, etc. Add them to `internal/platform/linux.go`.
- **Windows support** — `internal/platform/windows.go` doesn't exist yet. The `Platform` interface makes it straightforward to add.
- **New scanners** — Go module cache, Android SDK, Terraform providers, pip download cache, etc. Each scanner is a small self-contained file in `internal/scan/`.

To add a new scanner, implement the `scan.Scanner` interface and register it in `internal/ui/model.go`'s `startScanning()`:

```go
type Scanner interface {
    Name() string
    Scan(ctx context.Context, roots []platform.Root, sc ScanContext, out chan<- Item)
}
```

---

## License

MIT — see [LICENSE](LICENSE).

---

*Built by [Lemuel Okoli](https://github.com/okolilemuel). If broom saved you disk space, consider giving it a ⭐.*
