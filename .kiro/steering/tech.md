# Tech Stack

## Language & Runtime

- Go 1.21
- Module: `github.com/petercb/file-updater`

## Key Dependencies

| Library | Purpose |
|---------|---------|
| `github.com/hashicorp/go-getter` | Multi-protocol file downloading |
| `oras.land/oras-go/v2` | OCI registry pull via ORAS |
| `github.com/cpuguy83/dockercfg` | Docker credential resolution |
| `github.com/fsnotify/fsnotify` | File system watching |
| `gopkg.in/yaml.v3` | YAML config parsing |

## Build & Release

- **Build tool**: [GoReleaser](https://goreleaser.com/) (`.goreleaser.yaml`)
- **CI**: CircleCI (`.circleci/config.yml`) — Go 1.21 executor
- **Container**: Built with `ko` via GoReleaser, published to `ghcr.io/petercb`
- **Platforms**: darwin/linux/windows, amd64/arm64
- **Build tag**: `static_build` with `CGO_ENABLED=0`

## Linting & Quality

- **golangci-lint** (`.golangci.yaml`) — run via pre-commit hook
- **yamllint** — YAML file validation
- **markdownlint-cli2** — Markdown linting
- **codespell** — Spell checking
- **pre-commit** — Git hooks orchestration (`.pre-commit-config.yaml`)

## Common Commands

```sh
# Run tests
go test ./... -race -covermode=atomic

# Run linter
golangci-lint run

# Build locally
go build -tags static_build -o file-updater .

# Tidy modules
go mod tidy

# Run pre-commit hooks
pre-commit run --all-files

# Release (CI uses this)
curl -sfL https://goreleaser.com/static/run | bash
```
