package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDomainEntry_UnmarshalYAML_LegacyFormat(t *testing.T) {
	yaml := `
domains:
  backend:
    - api
    - worker
  frontend:
    - web-app
`
	cfg := loadFromString(t, yaml)

	be, ok := cfg.Domains["backend"]
	if !ok {
		t.Fatal("expected backend domain in config")
	}
	if len(be.Repos) != 2 || be.Repos[0] != "api" || be.Repos[1] != "worker" {
		t.Errorf("backend repos = %v, want [api worker]", be.Repos)
	}
	if len(be.Extensions) != 0 {
		t.Errorf("backend extensions = %v, want empty", be.Extensions)
	}

	fe, ok := cfg.Domains["frontend"]
	if !ok {
		t.Fatal("expected frontend domain in config")
	}
	if len(fe.Repos) != 1 || fe.Repos[0] != "web-app" {
		t.Errorf("frontend repos = %v, want [web-app]", fe.Repos)
	}
}

func TestDomainEntry_UnmarshalYAML_NewFormat(t *testing.T) {
	yaml := `
domains:
  mobile:
    repos:
      - ios-app
      - android-app
    extensions:
      - .swift
      - .kt
`
	cfg := loadFromString(t, yaml)

	m, ok := cfg.Domains["mobile"]
	if !ok {
		t.Fatal("expected mobile domain in config")
	}
	if len(m.Repos) != 2 || m.Repos[0] != "ios-app" {
		t.Errorf("mobile repos = %v, want [ios-app android-app]", m.Repos)
	}
	if len(m.Extensions) != 2 || m.Extensions[0] != ".swift" || m.Extensions[1] != ".kt" {
		t.Errorf("mobile extensions = %v, want [.swift .kt]", m.Extensions)
	}
}

func TestDomainEntry_UnmarshalYAML_ExtensionsOnly(t *testing.T) {
	yaml := `
domains:
  data:
    extensions:
      - .ipynb
      - .py
`
	cfg := loadFromString(t, yaml)

	d, ok := cfg.Domains["data"]
	if !ok {
		t.Fatal("expected data domain in config")
	}
	if len(d.Repos) != 0 {
		t.Errorf("data repos = %v, want empty", d.Repos)
	}
	if len(d.Extensions) != 2 || d.Extensions[0] != ".ipynb" {
		t.Errorf("data extensions = %v, want [.ipynb .py]", d.Extensions)
	}
}

func TestDomainEntry_UnmarshalYAML_MixedFormats(t *testing.T) {
	yaml := `
domains:
  backend:
    - api
    - worker
  mobile:
    repos:
      - ios-app
    extensions:
      - .swift
  data:
    extensions:
      - .ipynb
`
	cfg := loadFromString(t, yaml)

	if len(cfg.Domains) != 3 {
		t.Fatalf("domains count = %d, want 3", len(cfg.Domains))
	}

	be := cfg.Domains["backend"]
	if len(be.Repos) != 2 {
		t.Errorf("backend repos = %v, want 2 entries", be.Repos)
	}
	if len(be.Extensions) != 0 {
		t.Errorf("backend extensions should be empty for legacy format")
	}

	mob := cfg.Domains["mobile"]
	if len(mob.Repos) != 1 || mob.Repos[0] != "ios-app" {
		t.Errorf("mobile repos = %v", mob.Repos)
	}
	if len(mob.Extensions) != 1 || mob.Extensions[0] != ".swift" {
		t.Errorf("mobile extensions = %v", mob.Extensions)
	}

	data := cfg.Domains["data"]
	if len(data.Repos) != 0 {
		t.Errorf("data repos should be empty")
	}
	if len(data.Extensions) != 1 {
		t.Errorf("data extensions = %v, want 1 entry", data.Extensions)
	}
}

func TestDomainEntry_UnmarshalYAML_BuiltInWithExtensions(t *testing.T) {
	yaml := `
domains:
  backend:
    extensions:
      - .ts
    repos:
      - api
`
	cfg := loadFromString(t, yaml)

	be := cfg.Domains["backend"]
	if len(be.Repos) != 1 || be.Repos[0] != "api" {
		t.Errorf("backend repos = %v, want [api]", be.Repos)
	}
	if len(be.Extensions) != 1 || be.Extensions[0] != ".ts" {
		t.Errorf("backend extensions = %v, want [.ts]", be.Extensions)
	}
}

func TestLoad_NoDomains(t *testing.T) {
	yaml := `
tau: 180
`
	cfg := loadFromString(t, yaml)

	if cfg.Domains == nil {
		// nil map is acceptable — it means no domain overrides
		return
	}
	if len(cfg.Domains) != 0 {
		t.Errorf("domains should be empty when not configured, got %v", cfg.Domains)
	}
}

func TestLoad_EmptyDomains(t *testing.T) {
	yaml := `
domains: {}
`
	cfg := loadFromString(t, yaml)

	if len(cfg.Domains) != 0 {
		t.Errorf("domains should be empty, got %v", cfg.Domains)
	}
}

func TestUseDefaultDomains_NotSet(t *testing.T) {
	yaml := `
tau: 180
`
	cfg := loadFromString(t, yaml)
	if !cfg.UseDefaultDomains() {
		t.Error("UseDefaultDomains() should be true when not set")
	}
}

func TestUseDefaultDomains_True(t *testing.T) {
	yaml := `
default_domains: true
`
	cfg := loadFromString(t, yaml)
	if !cfg.UseDefaultDomains() {
		t.Error("UseDefaultDomains() should be true when explicitly true")
	}
}

func TestUseDefaultDomains_False(t *testing.T) {
	yaml := `
default_domains: false
domains:
  server:
    extensions:
      - .go
      - .py
  client:
    extensions:
      - .ts
      - .tsx
`
	cfg := loadFromString(t, yaml)
	if cfg.UseDefaultDomains() {
		t.Error("UseDefaultDomains() should be false when set to false")
	}
	if len(cfg.Domains) != 2 {
		t.Fatalf("domains count = %d, want 2", len(cfg.Domains))
	}
	server := cfg.Domains["server"]
	if len(server.Extensions) != 2 || server.Extensions[0] != ".go" {
		t.Errorf("server extensions = %v, want [.go .py]", server.Extensions)
	}
}

// loadFromString writes YAML to a temp file and loads it as Config.
func loadFromString(t *testing.T, yamlContent string) *Config {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "eis.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load(%s): %v", path, err)
	}
	return cfg
}
