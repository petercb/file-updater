# Product: file-updater

A Go CLI utility that downloads files from remote sources to local destinations based on a YAML configuration file.

## Core Behavior

- Reads a YAML config listing source/destination pairs
- Downloads files using [go-getter](https://pkg.go.dev/github.com/hashicorp/go-getter) (supports git, S3, GCS, HTTP/S, file, hg)
- Adds OCI/ORAS registry support (ghcr.io, ECR, GCR, Azure ACR, GitLab, Quay, localhost)
- By default watches the config file for changes and re-fetches on modification
- `--fetch-and-exit` flag for one-shot mode (no file watching)

## Primary Use Case

Kubernetes sidecar container that updates files (e.g., Minecraft server plugins) whenever a ConfigMap changes.

## Repository

- Owner: petercb
- Source: github.com/petercb/file-updater
- Container image: ghcr.io/petercb/file-updater
