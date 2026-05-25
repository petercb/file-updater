// Package main implements the file-updater CLI, which downloads files from
// remote sources to local destinations based on a YAML configuration file.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/fsnotify/fsnotify"
	"github.com/petercb/file-updater/internal/downloader"
	"gopkg.in/yaml.v3"
)

type File struct {
	Source string
	Dest   string
}

type Config struct {
	Files []File
}

func main() {
	var runOnce bool
	flag.BoolVar(
		&runOnce,
		"fetch-and-exit",
		false,
		"Don't monitor the file(s), just evaluate, fetch once and then exit",
	)
	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Fatal("Fatal: Need to specify at least one config file to process!")
	}

	files := initialFetch(flag.Args())

	if !runOnce {
		watchFiles(files)
	}
}

func watchFiles(files []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating watcher:", err)
	}

	for _, absfile := range files {
		basedir := filepath.Dir(absfile)

		if err := watcher.Add(basedir); err != nil {
			watcher.Close() //nolint:errcheck // best-effort cleanup before exit
			log.Fatal("Error adding config directory to watcher:", err)
		}
		log.Printf("Watching %s for changes", absfile)
	}

	defer func() {
		if err := watcher.Close(); err != nil {
			log.Println("Error closing watcher:", err)
		}
	}()

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Create|fsnotify.Write) > 0 {
				if slices.Contains(files, event.Name) {
					log.Println("File modified:", event.Name)
					fetchFiles(event.Name)
				} else if filepath.Base(event.Name) == "..data" {
					// Kubernetes ConfigMap update detected. Kubelet atomically
					// replaces the ..data symlink when a ConfigMap changes.
					// The actual config file is a symlink chain through ..data,
					// so we re-fetch all configs in the affected directory.
					dir := filepath.Dir(event.Name)
					for _, f := range files {
						if filepath.Dir(f) == dir {
							log.Println("ConfigMap update detected, re-processing:", f)
							fetchFiles(f)
						}
					}
				}
			}
		case err := <-watcher.Errors:
			log.Println("Error watching file:", err)
		}
	}
}

func initialFetch(conffiles []string) (files []string) {
	for _, f := range conffiles {
		log.Println("Processing file: ", f)
		absfile, err := filepath.Abs(f)
		if err != nil {
			log.Fatal("Error resolving absolute path:", err)
		}
		files = append(files, absfile)
		fetchFiles(f)
	}
	return
}

func fetchFiles(configfile string) {
	log.Println("Parsing config file")
	ctx := context.Background()
	config := loadConfig(configfile)
	for _, f := range config.Files {
		log.Printf("Fetching file [%s] to [%s]", f.Source, f.Dest)
		if err := downloader.Download(ctx, f.Dest, f.Source); err != nil {
			log.Printf("Failed to fetch %s to %s: %v", f.Source, f.Dest, err)
		}
	}
	log.Println("Done")
}

func loadConfig(filePath string) Config {
	// Read the YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading YAML file:", err)
	}

	// Parse YAML data into the Config struct
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("Error unmarshalling YAML data:", err)
	}

	return config
}
