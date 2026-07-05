package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestRunGetExclusions(t *testing.T) {
	withConfigFile(t, "EXCLUSIONS=foo,bar\n")

	out := captureStdout(t, func() {
		if err := runGet(&cmd{args: []string{"ex"}}); err != nil {
			t.Fatalf("run get exclusions: %v", err)
		}
	})

	if strings.TrimSpace(out) != "EXCLUSIONS=foo, bar" {
		t.Fatalf("unexpected get output: %q", out)
	}
	if err := runGet(&cmd{args: []string{"unknown"}}); err == nil {
		t.Fatal("expected unknown get command error")
	}
}

func TestRunExAddAndRemove(t *testing.T) {
	withConfigFile(t, "EXCLUSIONS=foo\n")

	if err := runEx(&cmd{opt: map[string]string{"add": "bar"}}); err != nil {
		t.Fatalf("run ex add: %v", err)
	}
	content, err := os.ReadFile("mfeeder.conf")
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(content), "EXCLUSIONS=foo,bar") {
		t.Fatalf("add did not update config: %q", content)
	}

	if err := runEx(&cmd{opt: map[string]string{"rm": "foo"}}); err != nil {
		t.Fatalf("run ex rm: %v", err)
	}
	content, err = os.ReadFile("mfeeder.conf")
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(content), "EXCLUSIONS=bar") {
		t.Fatalf("remove did not update config: %q", content)
	}
}
