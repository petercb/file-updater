package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestLoadConfig(t *testing.T) {
	config := loadConfig("testdata/config.yaml")
	if len(config.Files) != 1 {
		t.Fatal("Incorrect number of elements")
	}
}

// setupConfigMapMount creates a simulated Kubernetes ConfigMap mount structure
// and returns the directory, the config symlink path, and the ..data link path.
func setupConfigMapMount(t *testing.T) (dir, configLink, dataLink string) {
	t.Helper()
	dir = t.TempDir()

	ts1 := filepath.Join(dir, "..2024_01_01_00_00_00.000000000")
	if err := os.MkdirAll(ts1, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte("files:\n  - source: https://example.com/v1.jar\n    dest: /tmp/\n")
	if err := os.WriteFile(filepath.Join(ts1, "config.yaml"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	dataLink = filepath.Join(dir, "..data")
	if err := os.Symlink(ts1, dataLink); err != nil {
		t.Fatal(err)
	}

	configLink = filepath.Join(dir, "config.yaml")
	if err := os.Symlink(filepath.Join("..data", "config.yaml"), configLink); err != nil {
		t.Fatal(err)
	}

	return dir, configLink, dataLink
}

// simulateConfigMapUpdate creates a new timestamped directory and atomically
// replaces the ..data symlink, mimicking kubelet's ConfigMap update behavior.
func simulateConfigMapUpdate(t *testing.T, dir, dataLink string) []byte {
	t.Helper()

	ts2 := filepath.Join(dir, "..2024_01_02_00_00_00.000000000")
	if err := os.MkdirAll(ts2, 0o755); err != nil {
		t.Fatal(err)
	}
	newContent := []byte("files:\n  - source: https://example.com/v2.jar\n    dest: /tmp/\n")
	if err := os.WriteFile(filepath.Join(ts2, "config.yaml"), newContent, 0o644); err != nil {
		t.Fatal(err)
	}

	tmpLink := filepath.Join(dir, "..data_tmp")
	if err := os.Symlink(ts2, tmpLink); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmpLink, dataLink); err != nil {
		t.Fatal(err)
	}

	return newContent
}

// TestConfigMapSymlinkDetection verifies that the watcher detects Kubernetes
// ConfigMap-style updates where the ..data symlink is atomically replaced.
// On Linux (inotify), the rename produces a Create event on "..data".
// On macOS (kqueue), it produces Create events on the config.yaml symlink itself.
// Either way, the watcher logic in watchFiles handles both paths.
func TestConfigMapSymlinkDetection(t *testing.T) {
	dir, configLink, dataLink := setupConfigMapMount(t)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := watcher.Close(); closeErr != nil {
			t.Logf("warning: watcher close: %v", closeErr)
		}
	}()

	if err = watcher.Add(dir); err != nil {
		t.Fatal(err)
	}

	// Let watcher settle before triggering the update
	time.Sleep(100 * time.Millisecond)

	newContent := simulateConfigMapUpdate(t, dir, dataLink)

	// The watcher should detect the update via one of two paths:
	// - Linux: CREATE event on "..data" (our new detection logic)
	// - macOS: CREATE event on "config.yaml" (existing logic)
	files := []string{configLink}
	detected := false
	timeout := time.After(2 * time.Second)

	for !detected {
		select {
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Create|fsnotify.Write) == 0 {
				continue
			}
			if event.Name == configLink {
				detected = true
			} else if filepath.Base(event.Name) == "..data" {
				evDir := filepath.Dir(event.Name)
				for _, f := range files {
					if filepath.Dir(f) == evDir {
						detected = true
					}
				}
			}
		case watchErr := <-watcher.Errors:
			t.Fatalf("watcher error: %v", watchErr)
		case <-timeout:
			t.Fatal("timed out waiting for ConfigMap update event")
		}
	}

	// Verify the config file now resolves to new content
	data, err := os.ReadFile(configLink)
	if err != nil {
		t.Fatalf("failed to read config through symlink: %v", err)
	}
	if string(data) != string(newContent) {
		t.Errorf("config content not updated:\n  got:  %q\n  want: %q", string(data), string(newContent))
	}
}
