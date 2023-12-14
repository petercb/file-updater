package main

import "testing"

func TestLoadConfig(t *testing.T) {
	config := loadConfig("testdata/config.yaml")
	if len(config.Files) != 1 {
		t.Fatal("Incorrect number of elements")
	}
}
