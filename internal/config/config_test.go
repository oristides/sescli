package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfigMatchesOperatorDefaults(t *testing.T) {
	cfg := Default()

	if cfg.DefaultPreset != "centro" {
		t.Fatalf("expected centro default preset, got %q", cfg.DefaultPreset)
	}
	if cfg.Audience != "adulto" || cfg.Profile != "cultural" || cfg.Format != "json" {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
	if len(cfg.Presets["centro"]) == 0 {
		t.Fatalf("expected seeded centro preset from zonacentral: %#v", cfg.Presets)
	}
}

func TestWriteAndLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := Default()
	cfg.DefaultPreset = "custom"
	cfg.Presets["custom"] = []string{"2", "43"}

	if err := Write(path, cfg, false); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DefaultPreset != "custom" || len(loaded.Presets["custom"]) != 2 {
		t.Fatalf("unexpected loaded config: %#v", loaded)
	}
}

func TestWriteRefusesOverwriteUnlessForced(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, Default(), false); err == nil {
		t.Fatalf("expected overwrite protection")
	}
	if err := Write(path, Default(), true); err != nil {
		t.Fatalf("force overwrite failed: %v", err)
	}
}

func TestEnsureCreatesDefaultConfigIfMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	created, err := Ensure(path)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatalf("expected config to be created")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}

	created, err = Ensure(path)
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatalf("second ensure should not overwrite existing config")
	}
}

func TestSetupInteractiveWritesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	input := strings.NewReader("centro\nadulto\ncinema\nwhatsapp\n15\n")
	var output bytes.Buffer

	cfg, err := Setup(input, &output, path, true)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Defaults.Where != "centro" || cfg.Defaults.What != "cinema" || cfg.Defaults.Format != "whatsapp" || cfg.Defaults.Limit != 15 {
		t.Fatalf("unexpected setup config: %#v", cfg)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Defaults.What != "cinema" {
		t.Fatalf("expected nested config to load, got %#v", loaded)
	}
	if !strings.Contains(output.String(), "Default venue") {
		t.Fatalf("expected prompt output, got %q", output.String())
	}
}

func TestLoadOldFlatConfigShape(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"default_preset":"centro","audience":"adulto","profile":"all","format":"pretty","limit":10,"page":2}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Defaults.Where != "centro" || cfg.Defaults.What != "all" || cfg.Defaults.Format != "pretty" || cfg.Defaults.Limit != 10 || cfg.Defaults.Page != 2 {
		t.Fatalf("old config was not migrated: %#v", cfg)
	}
}
