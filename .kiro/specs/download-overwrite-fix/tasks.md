# Implementation Plan

- [x] 1. Write bug condition exploration test
  - **Property 1: Bug Condition** — Download Does Not Overwrite Existing Destination
  - **CRITICAL**: This test MUST FAIL on unfixed code — failure confirms the bug exists
  - **DO NOT attempt to fix the test or the code when it fails**
  - **NOTE**: This test encodes the expected behavior — it will validate the fix when it passes after implementation
  - **GOAL**: Surface counterexamples that demonstrate go-getter does not overwrite existing destinations
  - **Scoped PBT Approach**: Scope the property to the concrete failing case — destination file already exists with stale content, download a different file to the same path
  - Create test file: `internal/downloader/downloader_test.go`
  - Use `t.TempDir()` for test isolation
  - Use a local `file://` source (via go-getter's FileGetter) to avoid network dependencies
  - Test steps:
    1. Create a source file with known content (e.g., "version-2") in a temp directory
    2. Create a destination file with different content (e.g., "version-1") simulating a prior download
    3. Call `downloader.Download(ctx, dst, sourceURL)` where `dst` already exists
    4. Assert destination content equals source content ("version-2")
  - Run test on UNFIXED code: `go test ./internal/downloader/ -race -v -run TestDownloadOverwrite`
  - **EXPECTED OUTCOME**: Test FAILS (destination still contains "version-1" — confirms the bug)
  - Document counterexample: `Download(ctx, existingFile, validSource)` does not update destination content
  - Mark task complete when test is written, run, and failure is documented
  - _Requirements: 1.1_

- [x] 2. Implement the fix
  - [x] 2.1 Add `os.RemoveAll(dst)` before `client.Get()` in `Download` function
    - File: `internal/downloader/downloader.go`
    - Add `"os"` to the import block
    - Insert `os.RemoveAll(dst)` call after successful `Detect` but before `getter.Client` construction
    - Add error handling: wrap removal failure with `fmt.Errorf("removing existing destination: %w", err)`
    - `os.RemoveAll` is a no-op when path does not exist (preserves first-time download behavior)
    - _Bug_Condition: isBugCondition(input) where pathExists(input.dst) AND isValidURL(input.url)_
    - _Expected_Behavior: After Download completes, dst contains fresh content from source_
    - _Preservation: First-time downloads, detection failures, and invalid URLs unchanged_
    - _Requirements: 2.1, 2.2, 3.1, 3.2, 3.3, 3.4_

  - [x] 2.2 Verify bug condition exploration test now passes
    - **Property 1: Expected Behavior** — Download Overwrites Existing Destination
    - **IMPORTANT**: Re-run the SAME test from task 1 — do NOT write a new test
    - The test from task 1 encodes the expected behavior (destination updated with fresh content)
    - Run: `go test ./internal/downloader/ -race -v -run TestDownloadOverwrite`
    - **EXPECTED OUTCOME**: Test PASSES (confirms bug is fixed)
    - _Requirements: 2.1, 2.2_

- [x] 3. Write fix verification and preservation tests
  - **Property 2: Preservation** — First-Time Downloads and Error Paths Unchanged
  - **IMPORTANT**: These tests verify the fix works AND preserves existing behavior
  - Add to `internal/downloader/downloader_test.go` using table-driven subtests
  - Use `t.TempDir()` for isolation, local `file://` sources where possible
  - Test cases:
    1. **First-time download**: Download to a non-existent destination path — verify file is created with correct content (preservation of requirement 3.1)
    2. **Directory overwrite**: Create a directory at `dst` with stale content, download new content — verify directory is replaced (fix verification for requirement 2.2)
    3. **Invalid URL detection failure**: Call `Download` with an invalid URL — verify error is returned and no filesystem changes occur (preservation of requirement 3.3)
    4. **Non-existent source**: Call `Download` with a valid but non-existent file path — verify error from `client.Get()` (preservation of requirement 3.2)
  - Run: `go test ./internal/downloader/ -race -v`
  - **EXPECTED OUTCOME**: All tests PASS (confirms fix works and no regressions)
  - _Requirements: 2.2, 3.1, 3.2, 3.3, 3.4_

- [x] 4. Checkpoint — Ensure all tests pass
  - Run full test suite: `go test ./... -race -covermode=atomic`
  - Run linter: `golangci-lint run`
  - Ensure all tests pass, ask the user if questions arise
