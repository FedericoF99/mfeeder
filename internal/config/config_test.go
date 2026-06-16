package config

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestDefaultConf(t *testing.T) {
	t.Cleanup(func() {
		os.Remove(filePath)
	})

	cfg, err := LoadConfig(true)
	if err != nil {
		t.Errorf("error loading config: %v", err)
	}

	for _, exc := range strings.Split(defaultExclusions, ",") {
		if !slices.Contains(cfg.exclusions, exc) {
			t.Errorf("exclusion %s not found in config", exc)
		}
	}
}

func TestFileConf(t *testing.T) {
	createTempConf(t, "foo,bar")
	cfg, err := LoadConfig(true)
	if err != nil {
		t.Errorf("error loading config: %v", err)
	}

	if !slices.Contains(cfg.exclusions, "foo") || !slices.Contains(cfg.exclusions, "bar") {
		t.Errorf("exclusions not found in config: %v", cfg.exclusions)
	}
}

func TestEmptyFileConf(t *testing.T) {
	createTempConf(t, "")
	cfg, err := LoadConfig(true)
	if err != nil {
		t.Errorf("error loading config: %v", err)
	}

	if len(cfg.exclusions) != 0 {
		t.Errorf("exclusions should be empty in this test: %v", cfg.exclusions)
	}
}

func TestRmExclusion(t *testing.T) {
	createTempConf(t, "foo,bar")
	err := RmExclusion("foo")
	if err != nil {
		t.Errorf("error removing exclusion: %v", err)
	}
	cfg, err := LoadConfig(true)
	if err != nil {
		t.Errorf("error loading config: %v", err)
	}
	if slices.Contains(cfg.exclusions, "foo") {
		t.Errorf("remove exclusion failed")
	}
}

func TestAddExclusion(t *testing.T) {
	createTempConf(t, "foo,bar")
	err := AddExclusion("go")
	if err != nil {
		t.Errorf("error adding exclusion: %v", err)
	}
	cfg, err := LoadConfig(true)
	if err != nil {
		t.Errorf("error loading config: %v", err)
	}
	if !slices.Contains(cfg.exclusions, "go") {
		t.Errorf("add exclusion failed")
	}
}

func createTempConf(t *testing.T, exclusions string) {
	dir := os.TempDir()
	path := dir + "/" + filePath
	err := os.WriteFile(path, []byte("EXCLUSIONS="+exclusions+"\n"), 0644)
	if err != nil {
		t.Errorf("error creating test file: %v", err)
	}
	filePath = "/tmp/" + filePath
	t.Cleanup(func() {
		filePath = strings.TrimPrefix(filePath, "/tmp/")
	})
}
