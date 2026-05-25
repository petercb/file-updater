# Download Overwrite Fix — Bugfix Design

## Overview

The `Download` function in `internal/downloader/downloader.go` fails to overwrite existing files or directories at the destination path on subsequent fetches. This is because `go-getter`'s `Client` with `Mode: getter.ClientModeAny` does not replace existing content. The fix adds an `os.RemoveAll(dst)` call after successful URL detection but before invoking `client.Get()`, ensuring fresh content is always written.

## Glossary

- **Bug_Condition (C)**: The destination path (`dst`) already exists (as a file or directory) when `Download` is called
- **Property (P)**: After `Download` completes successfully, the destination contains the latest content from the remote source, regardless of prior state
- **Preservation**: First-time downloads, error handling for invalid URLs, and detection failures must remain unchanged
- **Download**: The exported function in `internal/downloader/downloader.go` that orchestrates URL detection and go-getter invocation
- **Detect**: The exported function that validates and normalizes the source URL before download
- **go-getter**: HashiCorp library providing multi-protocol file downloading (`github.com/hashicorp/go-getter`)

## Bug Details

### Bug Condition

The bug manifests when `Download` is called with a `dst` path that already exists on the filesystem. The `go-getter` client silently skips writing when the destination is already present, so stale content persists.

**Formal Specification:**
```
FUNCTION isBugCondition(input)
  INPUT: input of type DownloadInput {dst: string, url: string}
  OUTPUT: boolean

  RETURN pathExists(input.dst)
         AND isValidURL(input.url)
         AND detectSucceeds(input.url, input.dst)
END FUNCTION
```

### Examples

- `Download(ctx, "/tmp/plugin.jar", "https://example.com/plugin-v2.jar")` where `/tmp/plugin.jar` already contains v1 → destination still contains v1 after call (bug)
- `Download(ctx, "/tmp/configs/", "git::https://github.com/org/configs")` where `/tmp/configs/` already exists → directory not updated (bug)
- `Download(ctx, "/tmp/new-file.jar", "https://example.com/file.jar")` where `/tmp/new-file.jar` does not exist → downloads correctly (not a bug)
- `Download(ctx, "/tmp/plugin.jar", "://invalid")` where detection fails → returns error, destination untouched (not a bug)

## Expected Behavior

### Preservation Requirements

**Unchanged Behaviors:**
- First-time downloads to non-existent destinations must succeed exactly as before
- Invalid or unreachable URLs must return an error without modifying the filesystem
- URL detection failures must return a detection error without modifying the filesystem
- The function signature, return type, and error wrapping patterns remain identical
- The `Detect` function is not modified

**Scope:**
All inputs where `dst` does NOT already exist, or where the URL is invalid/unreachable, should be completely unaffected by this fix. This includes:
- First-time downloads (destination does not exist)
- Invalid source URLs that fail detection
- Unreachable URLs that fail during `client.Get()`
- Any call where `Detect` returns an error

## Hypothesized Root Cause

Based on code analysis of `internal/downloader/downloader.go`:

1. **go-getter ClientModeAny behavior**: When `Mode` is `getter.ClientModeAny` and the destination already exists, go-getter does not overwrite the existing content. It either silently succeeds without writing or returns an error depending on the getter implementation.

2. **No pre-download cleanup**: The `Download` function creates a `getter.Client` and calls `client.Get()` without any preparation of the destination path. There is no removal or truncation of existing content before the download attempt.

3. **Design assumption mismatch**: The function was likely written assuming single-use downloads. The watch-mode re-fetch pattern (via fsnotify in `main.go`) calls `Download` repeatedly for the same destination, exposing this limitation.

## Correctness Properties

Property 1: Bug Condition — Overwrite on Re-download

_For any_ download input where the destination path already exists and the URL is valid and detectable, the fixed `Download` function SHALL remove the existing destination before invoking go-getter, resulting in the destination containing the freshly downloaded content.

**Validates: Requirements 2.1, 2.2**

Property 2: Preservation — Non-Existent Destination and Error Paths

_For any_ download input where the destination does NOT already exist, OR where the URL fails detection, the fixed `Download` function SHALL produce the same result as the original function, preserving first-time download behavior and error handling.

**Validates: Requirements 3.1, 3.2, 3.3, 3.4**

## Fix Implementation

### Changes Required

**File**: `internal/downloader/downloader.go`

**Function**: `Download`

**Specific Changes**:

1. **Add `os` import**: Add `"os"` to the import block (it is not currently imported in this file).

2. **Add `os.RemoveAll(dst)` after detection succeeds**: Insert a call to `os.RemoveAll(dst)` between the successful `Detect` call and the `getter.Client` construction. This placement ensures:
   - We do NOT remove the destination if URL detection fails (preserving requirement 3.3)
   - We DO remove the destination before go-getter attempts to write (fixing requirements 2.1, 2.2)
   - `os.RemoveAll` is a no-op if the path does not exist (preserving requirement 3.1)

3. **Add error handling for RemoveAll**: Wrap the `os.RemoveAll` call with error checking and return a wrapped error if removal fails.

4. **No other changes**: The `Detect` function, client construction, getters map, and detectors slice remain unchanged.

**Resulting code shape:**
```go
func Download(ctx context.Context, dst string, url string) error {
    opts := []getter.ClientOption{}

    detectedURL, err := Detect(url, dst)
    if err != nil {
        return fmt.Errorf("detecting url: %w", err)
    }

    // Remove existing destination to ensure fresh content on re-download.
    // os.RemoveAll is a no-op if dst does not exist.
    if err := os.RemoveAll(dst); err != nil {
        return fmt.Errorf("removing existing destination: %w", err)
    }

    client := &getter.Client{
        Ctx:       ctx,
        Src:       detectedURL,
        Dst:       dst,
        Pwd:       dst,
        Mode:      getter.ClientModeAny,
        Detectors: detectors,
        Getters:   getters,
        Options:   opts,
    }

    if err := client.Get(); err != nil {
        return fmt.Errorf("client get: %w", err)
    }

    return nil
}
```

## Testing Strategy

### Validation Approach

The testing strategy follows a two-phase approach: first, surface counterexamples that demonstrate the bug on unfixed code, then verify the fix works correctly and preserves existing behavior.

### Exploratory Bug Condition Checking

**Goal**: Surface counterexamples that demonstrate the bug BEFORE implementing the fix. Confirm that go-getter does not overwrite existing destinations.

**Test Plan**: Write a test that creates a file at the destination, then calls `Download` with a valid source URL pointing to different content. Assert that the destination content changes. Run on UNFIXED code to observe failure.

**Test Cases**:
1. **File overwrite test**: Create a file at `dst`, download a different file to same `dst` — assert content changed (will fail on unfixed code)
2. **Directory overwrite test**: Create a directory at `dst`, download a repo/archive to same `dst` — assert content changed (will fail on unfixed code)

**Expected Counterexamples**:
- Destination file retains original content after `Download` completes without error
- go-getter silently succeeds but does not write new content

### Fix Checking

**Goal**: Verify that for all inputs where the bug condition holds, the fixed function produces the expected behavior.

**Pseudocode:**
```
FOR ALL input WHERE isBugCondition(input) DO
  result := Download_fixed(input.ctx, input.dst, input.url)
  ASSERT result = nil
  ASSERT fileContent(input.dst) = remoteContent(input.url)
END FOR
```

### Preservation Checking

**Goal**: Verify that for all inputs where the bug condition does NOT hold, the fixed function produces the same result as the original function.

**Pseudocode:**
```
FOR ALL input WHERE NOT isBugCondition(input) DO
  ASSERT Download_original(input) = Download_fixed(input)
END FOR
```

**Testing Approach**: Table-driven tests using `testing.T` subtests (matching project conventions). Property-based testing is recommended for preservation checking because:
- It generates many test cases automatically across the input domain
- It catches edge cases that manual unit tests might miss
- It provides strong guarantees that behavior is unchanged for all non-buggy inputs

**Test Plan**: Observe behavior on UNFIXED code first for non-existent destinations and invalid URLs, then write tests capturing that behavior.

**Test Cases**:
1. **First-time download preservation**: Download to a non-existent path — verify it succeeds identically
2. **Invalid URL preservation**: Call with an invalid URL — verify detection error is returned without filesystem changes
3. **Detection failure preservation**: Call with a URL that fails detection — verify error returned, no filesystem modification

### Unit Tests

- Test `Download` with pre-existing file at destination (overwrite case)
- Test `Download` with pre-existing directory at destination (overwrite case)
- Test `Download` with non-existent destination (first-time case)
- Test `Download` with invalid URL (error case)
- Test that `os.RemoveAll` is not called when `Detect` fails

### Property-Based Tests

- Generate random valid file paths and verify overwrite behavior when destination exists
- Generate random invalid URLs and verify error handling is preserved
- Test across file and directory destination types

### Integration Tests

- Test full fetch cycle: download, modify config, re-fetch, verify updated content
- Test watch-mode re-fetch scenario end-to-end
