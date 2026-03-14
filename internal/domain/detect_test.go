package domain

import (
	"testing"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  Domain
	}{
		{"backend", Backend},
		{"Backend", Backend},
		{"frontend", Frontend},
		{"mobile", Domain("Mobile")},
		{"data-science", Domain("Data-science")},
		{"", Unknown},
	}
	for _, tt := range tests {
		got := NormalizeName(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBuildExtMap_DefaultsOnly(t *testing.T) {
	m := BuildExtMap(nil, true)

	if m[".go"] != Backend {
		t.Errorf(".go = %v, want Backend", m[".go"])
	}
	if m[".ts"] != Frontend {
		t.Errorf(".ts = %v, want Frontend", m[".ts"])
	}
	if m[".tf"] != Infra {
		t.Errorf(".tf = %v, want Infra", m[".tf"])
	}
	if m[".c"] != Firmware {
		t.Errorf(".c = %v, want Firmware", m[".c"])
	}
}

func TestBuildExtMap_CustomOverride(t *testing.T) {
	// Move .ts from Frontend to Backend
	m := BuildExtMap(map[string][]string{
		"backend": {".ts"},
	}, true)

	if m[".ts"] != Backend {
		t.Errorf(".ts = %v, want Backend (overridden)", m[".ts"])
	}
	// Other defaults unchanged
	if m[".go"] != Backend {
		t.Errorf(".go = %v, want Backend", m[".go"])
	}
	if m[".tsx"] != Frontend {
		t.Errorf(".tsx = %v, want Frontend (unchanged)", m[".tsx"])
	}
}

func TestBuildExtMap_CustomDomain(t *testing.T) {
	m := BuildExtMap(map[string][]string{
		"mobile": {".swift", ".kt"},
	}, true)

	if m[".swift"] != Domain("Mobile") {
		t.Errorf(".swift = %v, want Mobile", m[".swift"])
	}
	if m[".kt"] != Domain("Mobile") {
		t.Errorf(".kt = %v, want Mobile", m[".kt"])
	}
	// .kt was Backend by default — now overridden
	// Verify other defaults still work
	if m[".go"] != Backend {
		t.Errorf(".go = %v, want Backend", m[".go"])
	}
}

func TestBuildExtMap_NoDotPrefix(t *testing.T) {
	// Extensions without leading dot should be normalized
	m := BuildExtMap(map[string][]string{
		"mobile": {"swift"},
	}, true)
	if m[".swift"] != Domain("Mobile") {
		t.Errorf(".swift = %v, want Mobile (dot added automatically)", m[".swift"])
	}
}

func TestBuildExtMap_CaseInsensitive(t *testing.T) {
	m := BuildExtMap(map[string][]string{
		"mobile": {".Swift"},
	}, true)
	if m[".swift"] != Domain("Mobile") {
		t.Errorf(".swift = %v, want Mobile (case normalized)", m[".swift"])
	}
}

func TestDetectFromFiles_DefaultExtMap(t *testing.T) {
	files := []string{"main.go", "handler.go", "config.yaml"}
	got := DetectFromFiles(files, nil)
	if got != Backend {
		t.Errorf("DetectFromFiles(go files) = %v, want Backend", got)
	}
}

func TestDetectFromFiles_CustomExtMap(t *testing.T) {
	extMap := BuildExtMap(map[string][]string{
		"mobile": {".swift", ".kt"},
	}, true)

	files := []string{"App.swift", "Model.swift", "ViewController.swift"}
	got := DetectFromFiles(files, extMap)
	if got != Domain("Mobile") {
		t.Errorf("DetectFromFiles(swift files, custom) = %v, want Mobile", got)
	}
}

func TestDetectFromFiles_CustomOverrideChangesDetection(t *testing.T) {
	// .ts is Frontend by default. Override to Backend.
	extMap := BuildExtMap(map[string][]string{
		"backend": {".ts", ".tsx"},
	}, true)

	files := []string{"server.ts", "handler.ts", "types.tsx"}
	got := DetectFromFiles(files, extMap)
	if got != Backend {
		t.Errorf("DetectFromFiles(ts files, overridden to backend) = %v, want Backend", got)
	}
}

func TestDetectFromFiles_InfraSpecialCase(t *testing.T) {
	// YAML-only should not be Infra
	files := []string{"config.yaml", "values.yml"}
	got := DetectFromFiles(files, nil)
	if got == Infra {
		t.Errorf("YAML-only files should not be classified as Infra")
	}

	// With .tf files, should be Infra
	files = []string{"main.tf", "variables.tf", "config.yaml"}
	got = DetectFromFiles(files, nil)
	if got != Infra {
		t.Errorf("files with .tf should be Infra, got %v", got)
	}
}

func TestBuildExtMap_NoDefaults(t *testing.T) {
	// default_domains: false — only custom extensions
	m := BuildExtMap(map[string][]string{
		"server": {".go", ".py"},
		"client": {".ts", ".tsx"},
	}, false)

	// Custom mappings work
	if m[".go"] != Domain("Server") {
		t.Errorf(".go = %v, want Server", m[".go"])
	}
	if m[".ts"] != Domain("Client") {
		t.Errorf(".ts = %v, want Client", m[".ts"])
	}
	// Default mappings are NOT present
	if _, ok := m[".java"]; ok {
		t.Errorf(".java should not be in map when defaults disabled, got %v", m[".java"])
	}
	if _, ok := m[".tf"]; ok {
		t.Errorf(".tf should not be in map when defaults disabled, got %v", m[".tf"])
	}
	if _, ok := m[".c"]; ok {
		t.Errorf(".c should not be in map when defaults disabled, got %v", m[".c"])
	}
}

func TestBuildExtMap_NoDefaults_EmptyCustom(t *testing.T) {
	// default_domains: false with no custom extensions — empty map
	m := BuildExtMap(nil, false)
	if len(m) != 0 {
		t.Errorf("expected empty map, got %d entries", len(m))
	}
}

func TestDetectFromFiles_NoDefaults(t *testing.T) {
	// With defaults disabled and custom-only extMap, .go files detect as Server
	extMap := BuildExtMap(map[string][]string{
		"server": {".go"},
	}, false)

	files := []string{"main.go", "handler.go"}
	got := DetectFromFiles(files, extMap)
	if got != Domain("Server") {
		t.Errorf("DetectFromFiles(no defaults, .go→Server) = %v, want Server", got)
	}

	// .java files are unrecognized → Unknown
	files = []string{"Main.java", "App.java"}
	got = DetectFromFiles(files, extMap)
	if got != Unknown {
		t.Errorf("DetectFromFiles(no defaults, .java unrecognized) = %v, want Unknown", got)
	}
}

func TestSortDomains_BuiltInOrder(t *testing.T) {
	input := []Domain{Infra, Backend, Frontend, Firmware}
	got := SortDomains(input)
	want := []Domain{Backend, Frontend, Infra, Firmware}
	for i, d := range got {
		if d != want[i] {
			t.Errorf("SortDomains[%d] = %v, want %v", i, d, want[i])
		}
	}
}

func TestSortDomains_CustomAfterBuiltIn(t *testing.T) {
	input := []Domain{Domain("Mobile"), Backend, Domain("Data"), Frontend}
	got := SortDomains(input)
	want := []Domain{Backend, Frontend, Domain("Data"), Domain("Mobile")}
	for i, d := range got {
		if d != want[i] {
			t.Errorf("SortDomains[%d] = %v, want %v", i, d, want[i])
		}
	}
}

func TestSortDomains_UnknownLast(t *testing.T) {
	input := []Domain{Unknown, Backend, Domain("Mobile")}
	got := SortDomains(input)
	want := []Domain{Backend, Domain("Mobile"), Unknown}
	for i, d := range got {
		if d != want[i] {
			t.Errorf("SortDomains[%d] = %v, want %v", i, d, want[i])
		}
	}
}

func TestSortDomains_CustomAlphabetical(t *testing.T) {
	input := []Domain{Domain("Zebra"), Domain("Alpha"), Domain("Mobile")}
	got := SortDomains(input)
	want := []Domain{Domain("Alpha"), Domain("Mobile"), Domain("Zebra")}
	for i, d := range got {
		if d != want[i] {
			t.Errorf("SortDomains[%d] = %v, want %v", i, d, want[i])
		}
	}
}
