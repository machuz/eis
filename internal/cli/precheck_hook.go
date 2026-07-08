package cli

// eis precheck-hook — Claude Code PreToolUse hook handler, implemented IN the
// eis binary (no bash/jq/subprocess, zero deps, one resolver as the single
// source of truth). It closes the local loop: right before an agent writes a
// file, it injects that module's structural-debt context.
//
// Contract (verified against the current Claude Code hooks docs,
// https://code.claude.com/docs/en/hooks):
//
//	INPUT (stdin JSON), PreToolUse:
//	  { "hook_event_name":"PreToolUse", "tool_name":"Write",
//	    "cwd":"/abs/repo", "tool_input": { "file_path":"/abs/repo/pkg/x.go" } }
//	  (Edit/MultiEdit also carry tool_input.file_path.)
//
//	OUTPUT (stdout JSON) to INJECT CONTEXT WITHOUT BLOCKING — exit 0:
//	  { "hookSpecificOutput": {
//	      "hookEventName":"PreToolUse",
//	      "additionalContext":"..." } }
//	  permissionDecision is OMITTED, so the tool proceeds through the normal
//	  permission flow; we only add context.
//
// Behavior:
//   - notable module  -> print additionalContext, exit 0.
//   - clean / not-found -> print NOTHING (silent-on-healthy = token budget).
//   - missing index / unresolvable path / ANY error -> FAIL OPEN: print nothing,
//     exit 0. The hook must NEVER block a write.
//
// FIREWALL: never emits owner names. The index carries none, and the injected
// directives use only structural facts (tier / days / recommendation).

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/machuz/eis/v2/internal/metric"
	"github.com/machuz/eis/v2/internal/output"
)

// hookInput is the subset of the PreToolUse stdin payload we read.
type hookInput struct {
	HookEventName string `json:"hook_event_name"`
	ToolName      string `json:"tool_name"`
	Cwd           string `json:"cwd"`
	ToolInput     struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
}

// hookOutput is the PreToolUse stdout payload that injects context (no block).
type hookOutput struct {
	HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
}

type hookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

func runPrecheckHook(args []string) error {
	fs := flag.NewFlagSet("precheck-hook", flag.ExitOnError)
	indexPath := fs.String("index", "", "Path to write-index.json (default: <cwd>/.eis/write-index.json)")
	if err := fs.Parse(args); err != nil {
		// Even flag errors must not block a write.
		return nil
	}

	// Everything below is best-effort: any failure falls through to a silent
	// exit 0 (fail-open). The hook must never stop the agent from writing.
	raw, err := io.ReadAll(os.Stdin)
	if err != nil || len(raw) == 0 {
		return nil
	}
	var in hookInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil
	}
	if in.ToolInput.FilePath == "" {
		return nil // no file to check (e.g. a Bash tool call) — silent.
	}

	resolvedIndex := *indexPath
	if resolvedIndex == "" {
		base := in.Cwd
		if base == "" {
			base, _ = os.Getwd()
		}
		resolvedIndex = filepath.Join(base, defaultWriteIndexPath)
	}

	idx, err := loadWriteIndex(resolvedIndex)
	if err != nil {
		return nil // no/broken index -> fail open, silent.
	}

	ctx, emit := precheckContext(in.ToolInput.FilePath, in.Cwd, idx)
	if !emit {
		return nil // clean or unresolved -> silent.
	}

	out := hookOutput{HookSpecificOutput: hookSpecificOutput{
		HookEventName:     "PreToolUse",
		AdditionalContext: ctx,
	}}
	// Best-effort emit; a write error here still must not block.
	_ = json.NewEncoder(os.Stdout).Encode(out)
	return nil
}

// loadWriteIndex reads and parses a write-index.json file.
func loadWriteIndex(path string) (output.WriteIndex, error) {
	var idx output.WriteIndex
	b, err := os.ReadFile(path)
	if err != nil {
		return idx, err
	}
	if err := json.Unmarshal(b, &idx); err != nil {
		return idx, err
	}
	return idx, nil
}

// resolveIndexModule resolves filePath → module id using the SAME patterns the
// index was built with (single source of truth), then looks the module up in the
// index. Returns (module, entry, true) only when the path resolves to a real,
// indexed module. Shared by precheck-hook and get-write-context so path→module
// resolution and lookup can never drift between them.
func resolveIndexModule(filePath, cwd string, idx output.WriteIndex) (string, output.WriteIndexModule, bool) {
	rel := repoRelPath(filePath, cwd)
	if rel == "" {
		return "", output.WriteIndexModule{}, false
	}
	// Build the resolver from the index's own module_patterns so path→module
	// resolution can never drift from how the index keys were produced.
	resolver := metric.NewModuleResolver(idx.ModulePatterns)
	mod := resolver.ModuleOf(rel)
	if mod == "" || mod == metric.PeripheralModule {
		return "", output.WriteIndexModule{}, false
	}
	entry, ok := idx.Modules[mod]
	if !ok {
		return mod, output.WriteIndexModule{}, false
	}
	return mod, entry, true
}

// precheckContext is the pure core for the hook: resolve + look up the module,
// then render its directive. Returns (context, true) only when the module is
// notable; ("", false) for clean / unresolved / not-in-index.
func precheckContext(filePath, cwd string, idx output.WriteIndex) (string, bool) {
	mod, entry, ok := resolveIndexModule(filePath, cwd, idx)
	if !ok {
		return "", false
	}
	directive := directiveFor(entry)
	if directive == "" {
		return "", false // clean module -> silent.
	}
	// One compact line per file (this rides on every write). Names the module so
	// the agent knows which write it applies to; carries NO owner identity.
	return fmt.Sprintf("[eis] %s — %s", mod, directive), true
}

// directiveFor maps a module's recommendation to a compact agent directive,
// interpolating only structural facts (never owner identity).
func directiveFor(m output.WriteIndexModule) string {
	switch m.Recommendation {
	case "orphaned_module":
		return fmt.Sprintf("Module has no active owner (left %dd ago, untouched %dd). "+
			"Prefer a minimal diff; add a test to establish shared ownership; "+
			"don't expand the public surface; flag to a human.",
			m.OwnerLeftDays, m.UntouchedDays)
	case "dead_module":
		return fmt.Sprintf("Module appears abandoned (no active owner, untouched %dd). "+
			"Consider deletion over extension; if extending, escalate the revive "+
			"decision to a human.", m.UntouchedDays)
	case "bus_factor_1":
		return "Single-owner module (bus-factor 1). Avoid concentrating more here; " +
			"write in a way others can build on."
	default:
		return "" // clean -> no output.
	}
}

// recommendationDirectives is the structured (array) form of the same guidance
// directiveFor renders as one prose line — used by get-write-context, which
// returns a query response rather than a hook envelope. Kept adjacent to
// directiveFor so the two stay consistent (same recommendation vocabulary). A
// clean module returns an empty (non-nil) slice. No owner identity, ever.
func recommendationDirectives(recommendation string) []string {
	switch recommendation {
	case "orphaned_module":
		return []string{
			"Prefer a minimal diff",
			"Add a test to establish shared ownership",
			"Don't expand the public surface",
			"Flag to a human",
		}
	case "dead_module":
		return []string{
			"Consider deletion over extension",
			"Escalate the revive decision to a human",
		}
	case "bus_factor_1":
		return []string{
			"Avoid concentrating more logic here",
			"Write so others can build on it",
		}
	default:
		return []string{}
	}
}

// repoRelPath returns filePath relative to the repo root (cwd), in forward-slash
// form suitable for ModuleResolver.ModuleOf. Returns "" when the path can't be
// placed inside the repo (fail-open).
func repoRelPath(filePath, cwd string) string {
	if filePath == "" {
		return ""
	}
	p := filePath
	if filepath.IsAbs(p) {
		base := cwd
		if base == "" {
			base, _ = os.Getwd()
		}
		r, err := filepath.Rel(base, p)
		if err != nil {
			return ""
		}
		p = r
	}
	p = filepath.ToSlash(p)
	p = strings.TrimPrefix(p, "./")
	if p == ".." || strings.HasPrefix(p, "../") {
		return "" // outside the repo — can't resolve a module id.
	}
	return p
}
