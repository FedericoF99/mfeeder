package config

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestDefaultConf(t *testing.T) {
	withConfigPath(t, filepath.Join(t.TempDir(), "mfeeder.conf"))

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
	withConfigContent(t, "EXCLUSIONS="+exclusions+"\n")
}

func TestLoadConfigWithoutOverwriteReturnsMissingFileError(t *testing.T) {
	withConfigPath(t, filepath.Join(t.TempDir(), "missing.conf"))

	if _, err := LoadConfig(false); err == nil {
		t.Fatal("expected missing config error")
	}
}

func TestAddExclusionDuplicateFails(t *testing.T) {
	withConfigContent(t, "EXCLUSIONS=foo,bar\n")

	if err := AddExclusion("foo"); err == nil {
		t.Fatal("expected duplicate exclusion error")
	}
}

func TestAddExclusionRejectsEmptyValue(t *testing.T) {
	withConfigContent(t, "EXCLUSIONS=foo,bar\n")

	if err := AddExclusion(""); err == nil {
		t.Fatal("expected empty exclusion error")
	}
}

func TestAddExclusionTrimsAndDetectsDuplicate(t *testing.T) {
	withConfigContent(t, "EXCLUSIONS=foo,bar\n")

	if err := AddExclusion(" foo "); err == nil {
		t.Fatal("expected spaced duplicate exclusion error")
	}
}

func TestRemoveMissingExclusionFails(t *testing.T) {
	withConfigContent(t, "EXCLUSIONS=foo,bar\n")

	if err := RmExclusion("missing"); err == nil {
		t.Fatal("expected missing exclusion error")
	}
}

func TestRemoveExclusionTrimsInput(t *testing.T) {
	withConfigContent(t, "EXCLUSIONS=foo,bar\n")

	if err := RmExclusion(" foo "); err != nil {
		t.Fatalf("remove exclusion should trim input: %v", err)
	}
	cfg, err := LoadConfig(false)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if slices.Contains(cfg.Exclusions(), "foo") {
		t.Fatalf("trimmed exclusion was not removed: %#v", cfg.Exclusions())
	}
}

func TestAddAndRemoveExclusionFailWhenConfigHasNoExclusions(t *testing.T) {
	withConfigContent(t, "OTHER=value\n")

	if err := AddExclusion("foo"); err == nil {
		t.Fatal("expected add exclusion to fail without EXCLUSIONS entry")
	}
	if err := RmExclusion("foo"); err == nil {
		t.Fatal("expected remove exclusion to fail without EXCLUSIONS entry")
	}
}

func TestGetExclusions(t *testing.T) {
	withConfigContent(t, "EXCLUSIONS=foo,bar\n")

	exclusions, err := GetExclusions()
	if err != nil {
		t.Fatalf("get exclusions: %v", err)
	}
	if !reflect.DeepEqual(exclusions, []string{"foo", "bar"}) {
		t.Fatalf("unexpected exclusions: %#v", exclusions)
	}
}

func TestLoadConfigIgnoresEmptyAndUnknownLines(t *testing.T) {
	withConfigContent(t, "\nUNKNOWN=value\n  EXCLUSIONS=foo,bar  \n")

	cfg, err := LoadConfig(false)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !reflect.DeepEqual(cfg.Exclusions(), []string{"foo", "bar"}) {
		t.Fatalf("unexpected exclusions: %#v", cfg.Exclusions())
	}
}

func TestLoadConfigTrimsExclusionValues(t *testing.T) {
	withConfigContent(t, "EXCLUSIONS=foo, bar, baz\n")

	cfg, err := LoadConfig(false)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !reflect.DeepEqual(cfg.Exclusions(), []string{"foo", "bar", "baz"}) {
		t.Fatalf("unexpected exclusions: %#v", cfg.Exclusions())
	}
}

func TestLoadConfigRejectsEmptyExclusionTokens(t *testing.T) {
	withConfigContent(t, "EXCLUSIONS=foo,,bar\n")

	if cfg, err := LoadConfig(false); err != nil {
		t.Fatal("expected error on empty exclusion token")
	} else {
		if len(cfg.Exclusions()) != 2 {
			t.Fatalf("expected 2 exclusions, got %d", len(cfg.Exclusions()))
		}
	}
}

func TestLoadConfigRejectsDuplicateExclusionTokens(t *testing.T) {
	withConfigContent(t, "EXCLUSIONS=foo,bar,foo\n")

	if cfg, err := LoadConfig(false); err != nil {
		t.Fatal("unexpected error on duplicate exclusion token")
	} else if cfg == nil {
		t.Fatalf("expected config to be non-nil")
	} else {
		if len(cfg.Exclusions()) != 2 {
			t.Fatalf("expected 2 exclusions, got %d", len(cfg.Exclusions()))
		}
	}
}

func withConfigContent(t *testing.T, content string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "mfeeder.conf")
	withConfigPath(t, path)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func withConfigPath(t *testing.T, path string) {
	t.Helper()
	t.Setenv("MFEEDER_DATA_DIR", filepath.Dir(path))
}
