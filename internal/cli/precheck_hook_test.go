package cli

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/machuz/eis/v2/internal/output"
)

// withStdin feeds input on os.Stdin, runs fn, and returns what fn wrote to
// stdout (reusing captureStd for the stdout swap).
func withStdin(t *testing.T, input string, fn func()) string {
	t.Helper()
	origIn := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	go func() { _, _ = io.WriteString(w, input); _ = w.Close() }()
	defer func() { os.Stdin = origIn; _ = r.Close() }()
	out, _ := captureStd(t, fn)
	return out
}

// fixtureIndex is a write-index with one debt module (orphaned), one dead, one
// at-risk, and one clean module. module_patterns empty ⇒ ModuleResolver falls
// back to the built-in defaults, so paths resolve to the 2-component ids used
// as keys here.
func fixtureIndex() output.WriteIndex {
	return output.WriteIndex{
		GeneratedAt:          "2026-07-09T00:00:00Z",
		Commit:               "deadbeef",
		ModulePatterns:       nil, // -> DefaultModulePatterns fallback
		ModulePatternsSource: "builtin_default",
		Modules: map[string]output.WriteIndexModule{
			"svc/auth":    {DebtTier: "Orphaned", OwnerLeftDays: 243, UntouchedDays: 243, OwnershipConcentration: 0.9, Recommendation: "orphaned_module"},
			"legacy/vm":   {DebtTier: "Dead", UntouchedDays: 900, Recommendation: "dead_module"},
			"core/engine": {AtRisk: true, OwnerActive: true, OwnershipConcentration: 0.88, Recommendation: "bus_factor_1"},
			"core/api":    {Recommendation: ""}, // clean
		},
	}
}

func TestPrecheckContext_DebtModuleInjectsDirective(t *testing.T) {
	idx := fixtureIndex()

	cases := []struct {
		name       string
		path, cwd  string
		wantSubstr []string
	}{
		{"orphaned", "/repo/svc/auth/login.go", "/repo",
			[]string{"svc/auth", "no active owner", "left 243d", "untouched 243d", "flag to a human"}},
		{"dead", "/repo/legacy/vm/run.go", "/repo",
			[]string{"legacy/vm", "abandoned", "untouched 900d", "deletion over extension"}},
		{"bus_factor_1", "/repo/core/engine/core.go", "/repo",
			[]string{"core/engine", "bus-factor 1", "Avoid concentrating"}},
		{"relative path (no cwd)", "svc/auth/x.go", "",
			[]string{"svc/auth", "no active owner"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, emit := precheckContext(c.path, c.cwd, idx)
			if !emit {
				t.Fatalf("expected emit=true for %s", c.path)
			}
			for _, s := range c.wantSubstr {
				if !strings.Contains(got, s) {
					t.Errorf("context missing %q\n got: %s", s, got)
				}
			}
			// Compact: single line.
			if strings.Contains(got, "\n") {
				t.Errorf("directive must be one line, got:\n%s", got)
			}
		})
	}
}

// Firewall: even if a rogue owner name were present in the index, the injected
// directive template must never surface identity tokens.
func TestPrecheckContext_NoOwnerNames(t *testing.T) {
	idx := fixtureIndex()
	got, emit := precheckContext("/repo/svc/auth/login.go", "/repo", idx)
	if !emit {
		t.Fatal("expected emit")
	}
	for _, name := range []string{"tanaka", "Mark Erikson", "Lee Byron", "owner:", "@"} {
		if strings.Contains(got, name) {
			t.Errorf("directive leaked identity token %q: %s", name, got)
		}
	}
}

func TestPrecheckContext_CleanModuleIsSilent(t *testing.T) {
	idx := fixtureIndex()
	// core/api is clean.
	if got, emit := precheckContext("/repo/core/api/handler.go", "/repo", idx); emit {
		t.Errorf("clean module should be silent, got: %s", got)
	}
	// A module not in the index at all -> silent (fail-open).
	if got, emit := precheckContext("/repo/brand/new/feature.go", "/repo", idx); emit {
		t.Errorf("unknown module should be silent, got: %s", got)
	}
	// Path outside the repo -> silent.
	if _, emit := precheckContext("/elsewhere/x.go", "/repo", idx); emit {
		t.Error("path outside repo should be silent")
	}
	// Empty index -> silent.
	if _, emit := precheckContext("/repo/svc/auth/x.go", "/repo", output.WriteIndex{}); emit {
		t.Error("empty index should be silent")
	}
}

// End-to-end: feed a real PreToolUse payload on stdin, capture stdout, and check
// the exact hookSpecificOutput envelope. Also covers fail-open when the index
// file is absent.
func TestRunPrecheckHook_StdinToStdout(t *testing.T) {
	dir := t.TempDir()
	// Write a fixture index at <dir>/.eis/write-index.json.
	if err := os.MkdirAll(dir+"/.eis", 0o755); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(fixtureIndex())
	if err := os.WriteFile(dir+"/.eis/write-index.json", b, 0o644); err != nil {
		t.Fatal(err)
	}

	payload := `{"hook_event_name":"PreToolUse","tool_name":"Write","cwd":"` + dir +
		`","tool_input":{"file_path":"` + dir + `/svc/auth/login.go"}}`

	out := withStdin(t, payload, func() {
		if err := runPrecheckHook(nil); err != nil {
			t.Fatalf("runPrecheckHook returned error (must be nil/fail-open): %v", err)
		}
	})

	var got hookOutput
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("stdout is not valid hook JSON: %v\n%s", err, out)
	}
	if got.HookSpecificOutput.HookEventName != "PreToolUse" {
		t.Errorf("hookEventName = %q, want PreToolUse", got.HookSpecificOutput.HookEventName)
	}
	if !strings.Contains(got.HookSpecificOutput.AdditionalContext, "svc/auth") {
		t.Errorf("additionalContext missing module: %s", got.HookSpecificOutput.AdditionalContext)
	}
	// permissionDecision must NOT be present (context-only injection).
	if strings.Contains(out, "permissionDecision") {
		t.Errorf("must not set permissionDecision, got: %s", out)
	}

	// Clean module -> empty stdout.
	cleanPayload := `{"hook_event_name":"PreToolUse","tool_name":"Write","cwd":"` + dir +
		`","tool_input":{"file_path":"` + dir + `/core/api/h.go"}}`
	cleanOut := withStdin(t, cleanPayload, func() { _ = runPrecheckHook(nil) })
	if strings.TrimSpace(cleanOut) != "" {
		t.Errorf("clean module must produce no output, got: %q", cleanOut)
	}

	// No index file -> fail-open, empty stdout, no error.
	noIdxPayload := `{"hook_event_name":"PreToolUse","tool_name":"Write","cwd":"/nonexistent",` +
		`"tool_input":{"file_path":"/nonexistent/svc/auth/x.go"}}`
	failOpen := withStdin(t, noIdxPayload, func() {
		if err := runPrecheckHook(nil); err != nil {
			t.Fatalf("missing index must fail open, got err: %v", err)
		}
	})
	if strings.TrimSpace(failOpen) != "" {
		t.Errorf("missing index must be silent, got: %q", failOpen)
	}
}
