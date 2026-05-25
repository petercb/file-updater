# Project Structure

```
.
├── main.go                  # CLI entrypoint: config loading, file watching, orchestration
├── main_test.go             # Top-level integration tests
├── go.mod / go.sum          # Go module definition
├── internal/
│   ├── downloader/          # File download engine
│   │   ├── downloader.go    # Download() and Detect() — wires detectors and getters
│   │   ├── oci_detector.go  # Detects OCI registry URLs (ACR, GCR, GHCR, ECR, etc.)
│   │   ├── oci_detector_test.go
│   │   └── oci_getter.go    # Pulls artifacts from OCI registries via ORAS
│   ├── network/             # Network utilities
│   │   ├── network.go       # Hostname() and IsLoopback() helpers
│   │   └── network_test.go
│   └── registry/            # OCI registry client setup
│       ├── client.go        # SetupClient() — auth, user-agent, plainHTTP for localhost
│       ├── client_test.go
│       └── testdata/        # Test fixtures (docker config.json)
├── testdata/
│   └── config.yaml          # Example/test configuration file
├── .circleci/config.yml     # CI pipeline
├── .goreleaser.yaml         # Release configuration
├── .golangci.yaml           # Linter configuration
├── .pre-commit-config.yaml  # Pre-commit hooks
└── .kiro/steering/          # AI assistant steering rules
```

## Architecture Notes

- **Flat main**: All CLI logic lives in `main.go` (config parsing, flag handling, watcher loop)
- **internal/ packages**: Not importable externally — keeps API surface minimal
- **Adapter pattern**: `downloader` package extends go-getter with custom OCI detector and getter
- **Borrowed code**: OCI support adapted from [OPA ConfTest v0.47.0](https://github.com/open-policy-agent/conftest/tree/v0.47.0/downloader) — source comments note this
- **Test style**: Table-driven tests using `testing.T` subtests (no third-party test framework)
- **testdata/**: Convention for test fixture files co-located with packages that use them
