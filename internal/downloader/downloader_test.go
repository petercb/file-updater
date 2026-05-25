package downloader

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestDownloadOverwritesExistingFile verifies that Download overwrites a stale
// file inside the destination directory with fresh content from the source.
// go-getter v1.7.3 handles this natively without any workaround.
func TestDownloadOverwritesExistingFile(t *testing.T) {
	// Setup: create a source file with known "new" content.
	sourceDir := t.TempDir()
	sourceFile := filepath.Join(sourceDir, "artifact.txt")
	sourceContent := []byte("version-2")
	if err := os.WriteFile(sourceFile, sourceContent, 0o644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Setup: destination directory with a stale version of the same file.
	destDir := t.TempDir()
	staleFile := filepath.Join(destDir, "artifact.txt")
	if err := os.WriteFile(staleFile, []byte("version-1"), 0o644); err != nil {
		t.Fatalf("failed to create stale file: %v", err)
	}

	// Act
	ctx := context.Background()
	err := Download(ctx, destDir, sourceFile)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Assert: file content is updated
	got, err := os.ReadFile(filepath.Join(destDir, "artifact.txt"))
	if err != nil {
		t.Fatalf("failed to read result file: %v", err)
	}
	if string(got) != string(sourceContent) {
		t.Errorf("file not overwritten: got %q, want %q", string(got), string(sourceContent))
	}
}

// TestDownloadPreservesOtherFiles verifies that downloading a file into a
// shared directory does not destroy unrelated files already present.
func TestDownloadPreservesOtherFiles(t *testing.T) {
	// Setup: source file
	sourceDir := t.TempDir()
	sourceFile := filepath.Join(sourceDir, "plugin.jar")
	if err := os.WriteFile(sourceFile, []byte("plugin-v2"), 0o644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Setup: destination directory with an unrelated file and a stale target
	destDir := t.TempDir()
	unrelatedFile := filepath.Join(destDir, "other-plugin.jar")
	if err := os.WriteFile(unrelatedFile, []byte("should-survive"), 0o644); err != nil {
		t.Fatalf("failed to create unrelated file: %v", err)
	}
	staleFile := filepath.Join(destDir, "plugin.jar")
	if err := os.WriteFile(staleFile, []byte("plugin-v1"), 0o644); err != nil {
		t.Fatalf("failed to create stale file: %v", err)
	}

	// Act
	ctx := context.Background()
	err := Download(ctx, destDir, sourceFile)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Assert: target file is updated
	got, err := os.ReadFile(filepath.Join(destDir, "plugin.jar"))
	if err != nil {
		t.Fatalf("target file missing: %v", err)
	}
	if string(got) != "plugin-v2" {
		t.Errorf("target not updated: got %q, want %q", string(got), "plugin-v2")
	}

	// Assert: unrelated file is preserved
	other, err := os.ReadFile(unrelatedFile)
	if err != nil {
		t.Fatalf("unrelated file was deleted: %v", err)
	}
	if string(other) != "should-survive" {
		t.Errorf("unrelated file content changed: got %q", string(other))
	}
}

// TestDownloadFirstTime verifies that downloading to a non-existent
// destination directory works correctly.
func TestDownloadFirstTime(t *testing.T) {
	// Setup: source file
	sourceDir := t.TempDir()
	sourceFile := filepath.Join(sourceDir, "source.txt")
	if err := os.WriteFile(sourceFile, []byte("fresh-content"), 0o644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Destination does not exist yet
	destDir := filepath.Join(t.TempDir(), "new-dest")

	// Act
	ctx := context.Background()
	err := Download(ctx, destDir, sourceFile)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Assert
	got, err := os.ReadFile(filepath.Join(destDir, "source.txt"))
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if string(got) != "fresh-content" {
		t.Errorf("file content = %q, want %q", string(got), "fresh-content")
	}
}

// TestDownloadErrorCases verifies that invalid inputs return errors
// without modifying the filesystem.
func TestDownloadErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		sourceURL string
	}{
		{
			name:      "invalid URL returns detection error",
			sourceURL: "://invalid",
		},
		{
			name:      "non-existent source returns error",
			sourceURL: "/tmp/this-path-definitely-does-not-exist-abc123/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := filepath.Join(t.TempDir(), "should-not-exist")
			ctx := context.Background()

			err := Download(ctx, dst, tt.sourceURL)
			if err == nil {
				t.Fatal("expected error but got nil")
			}

			// Destination should not be created on error
			if _, statErr := os.Stat(dst); !errors.Is(statErr, os.ErrNotExist) {
				t.Errorf("expected dst to not exist after error, stat returned: %v", statErr)
			}
		})
	}
}
