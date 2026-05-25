---
inclusion: auto
description: Project-specific patterns, preferences, and lessons learned over time (user-editable)
---

# Lessons Learned

This file captures project-specific patterns, coding preferences, common pitfalls, and architectural decisions that emerge during development. It serves as a workaround for continuous learning by allowing you to document patterns manually.

**How to use this file:**
1. The `extract-patterns` hook will suggest patterns after agent sessions
2. Review suggestions and add genuinely useful patterns below
3. Edit this file directly to capture team conventions
4. Keep it focused on project-specific insights, not general best practices

---

## Project-Specific Patterns

*Document patterns unique to this project that the team should follow.*

### Example: API Error Handling
```typescript
// Always use our custom ApiError class for consistent error responses
throw new ApiError(404, 'Resource not found', { resourceId });
```

---

## Code Style Preferences

*Document team preferences that go beyond standard linting rules.*

### Example: Import Organization
```typescript
// Group imports: external, internal, types
import { useState } from 'react';
import { Button } from '@/components/ui';
import type { User } from '@/types';
```

---

## Kiro Hooks

### `install.sh` is additive-only — it won't update existing installations
The installer skips any file that already exists in the target (`if [ ! -f ... ]`). Running it against a folder that already has `.kiro/` will not overwrite or update hooks, agents, or steering files. To push updates to an existing project, manually copy the changed files or remove the target files first before re-running the installer.

### README.md mirrors hook configurations — keep them in sync
The hooks table and Example 5 in README.md document the action type (`runCommand` vs `askAgent`) and behavior of each hook. When changing a hook's `then.type` or behavior, update both the hook file and the corresponding README entries to avoid misleading documentation.

### Prefer `askAgent` over `runCommand` for file-event hooks
`runCommand` hooks on `fileEdited` or `fileCreated` events spawn a new terminal session every time they fire, creating friction. Use `askAgent` instead so the agent handles the task inline. Reserve `runCommand` for `userTriggered` hooks where a manual, isolated terminal run is intentional (e.g., `quality-gate`).

---

## Common Pitfalls

*Document mistakes that have been made and how to avoid them.*

### go-getter `ClientModeAny` treats `dst` as a directory
When using `github.com/hashicorp/go-getter` with `ClientModeAny`, the destination path is treated as a **directory** — the downloaded file is placed inside it (e.g., `dst/filename`). It does NOT write directly to `dst` as a file path. Tests must assert on `filepath.Join(dst, filename)`, not on `dst` itself.

### go-getter v1.7.3 DOES overwrite files inside existing directories
Despite initial assumptions, go-getter v1.7.3 with `ClientModeAny` successfully overwrites files inside an existing directory destination. Do NOT use `os.RemoveAll(dst)` as a workaround — it destroys unrelated files in shared directories. Always verify actual library behavior with a direct test before adding workarounds.

### Kubernetes ConfigMap mounts use symlinks — fsnotify may miss updates
Kubernetes mounts ConfigMaps via symlinks (`..data` → `..2024_01_01/`). On update, the symlink target changes atomically. `fsnotify` may not fire a `Write` event on the config file path — it fires on the symlink directory. When debugging "file not updating" in K8s sidecars, investigate the fsnotify/symlink interaction before assuming the download layer is broken.

### fsnotify behavior differs between Linux and macOS for symlink renames
On Linux (inotify), atomically renaming a symlink (e.g., `..data`) produces a `CREATE` event on `..data` but NOT on files that symlink through it. On macOS (kqueue), the same operation produces `CREATE` events on the downstream symlinks (e.g., `config.yaml`). Tests for symlink-based file watching must handle both paths to pass cross-platform.

### `os.RemoveAll` placement matters for download safety
When adding pre-download cleanup (like `os.RemoveAll(dst)`), always place it AFTER URL validation/detection succeeds but BEFORE the actual download. This ensures we don't destroy existing content if the source URL is invalid.

### Example: Database Transactions
- Always wrap multiple database operations in a transaction
- Remember to handle rollback on errors
- Don't forget to close connections in finally blocks

---

## Architecture Decisions

*Document key architectural decisions and their rationale.*

### Example: State Management
- **Decision**: Use Zustand for global state, React Context for component trees
- **Rationale**: Zustand provides better performance and simpler API than Redux
- **Trade-offs**: Less ecosystem tooling than Redux, but sufficient for our needs

---

## Notes

- Keep entries concise and actionable
- Remove patterns that are no longer relevant
- Update patterns as the project evolves
- Focus on what's unique to this project
