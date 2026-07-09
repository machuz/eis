package cli

import (
	"strings"
	"testing"

	"github.com/machuz/eis/v2/internal/output"
)

// stubDigest returns a fixed digest so buildAnchors tests need no filesystem.
func stubDigest(abs string) (string, [2]int) {
	return "package x\nfunc H() {}", [2]int{1, 2}
}

func TestBuildAnchors_RanksBySurvivalTimesGravityAndGatesContested(t *testing.T) {
	stats := []AnchorStat{
		// module svc/order:
		// high survival + high mass, contested -> top anchor.
		{Module: "svc/order", File: "svc/order/order.go", AbsPath: "/r/svc/order/order.go", DecayedMass: 90, Lines: 100, Contributors: 4},
		// lower mass -> ranked below.
		{Module: "svc/order", File: "svc/order/util.go", AbsPath: "/r/svc/order/util.go", DecayedMass: 20, Lines: 40, Contributors: 3},
		// single-contributor -> excluded by the robust gate even though heavy.
		{Module: "svc/order", File: "svc/order/solo.go", AbsPath: "/r/svc/order/solo.go", DecayedMass: 200, Lines: 210, Contributors: 1},
		// another module, contested -> its own anchor list.
		{Module: "svc/pay", File: "svc/pay/pay.go", AbsPath: "/r/svc/pay/pay.go", DecayedMass: 50, Lines: 60, Contributors: 2},
	}

	r := buildAnchors(stats, 3, false, false, stubDigest)

	if len(r.Modules) != 2 {
		t.Fatalf("modules = %d, want 2 (svc/order, svc/pay)", len(r.Modules))
	}
	// Modules sorted by name.
	if r.Modules[0].Module != "svc/order" || r.Modules[1].Module != "svc/pay" {
		t.Fatalf("module order = %v", []string{r.Modules[0].Module, r.Modules[1].Module})
	}

	order := r.Modules[0]
	// solo.go must be gated out (single contributor); 2 anchors remain.
	if len(order.Anchors) != 2 {
		t.Fatalf("svc/order anchors = %d, want 2 (solo.go gated)", len(order.Anchors))
	}
	if order.Anchors[0].File != "svc/order/order.go" {
		t.Errorf("top anchor = %s, want svc/order/order.go (higher survival×gravity)", order.Anchors[0].File)
	}
	for _, a := range order.Anchors {
		if a.File == "svc/order/solo.go" {
			t.Error("single-contributor file must not be an anchor")
		}
	}

	top := order.Anchors[0]
	if top.ContestedByN != 4 {
		t.Errorf("contested_by_n = %d, want 4", top.ContestedByN)
	}
	if top.Survival != 0.9 { // 90/100
		t.Errorf("survival = %v, want 0.9", top.Survival)
	}
	if top.Gravity != 90 {
		t.Errorf("gravity = %v, want 90", top.Gravity)
	}
	if top.LineRange != [2]int{1, 2} || top.Digest == "" {
		t.Errorf("digest/line_range not wired: %+v", top)
	}
}

func TestBuildAnchors_TopNCap(t *testing.T) {
	var stats []AnchorStat
	for _, f := range []string{"a", "b", "c", "d", "e"} {
		stats = append(stats, AnchorStat{Module: "m", File: "m/" + f + ".go", DecayedMass: 10, Lines: 10, Contributors: 2})
	}
	r := buildAnchors(stats, 2, false, false, stubDigest)
	if len(r.Modules) != 1 || len(r.Modules[0].Anchors) != 2 {
		t.Fatalf("top=2 not honored: %+v", r.Modules)
	}
}

// Firewall: no author identity in the serialized anchors report.
func TestBuildAnchors_NoOwnerNames(t *testing.T) {
	stats := []AnchorStat{
		{Module: "svc/order", File: "svc/order/order.go", DecayedMass: 90, Lines: 100, Contributors: 4},
	}
	r := buildAnchors(stats, 3, false, false, func(abs string) (string, [2]int) {
		// Even a digest with a name-shaped token: firewall is about the STRUCT
		// fields, and contested_by_n is a count. (Digest is code, not authorship.)
		return "func H() {}", [2]int{1, 1}
	})
	var buf strings.Builder
	if err := output.EncodeAnchorsJSON(&buf, r); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, tok := range []string{"author", "owner", "top_author", "contributor\"", "\"Alice\"", "email"} {
		if strings.Contains(s, tok) {
			t.Errorf("anchors leaked identity-shaped field %q:\n%s", tok, s)
		}
	}
	// It DOES carry the anonymous count + weights.
	if !strings.Contains(s, `"contested_by_n": 4`) || !strings.Contains(s, `"gravity"`) {
		t.Errorf("expected anonymous count + weights:\n%s", s)
	}
}

func moduleByName(r output.AnchorsReport, name string) (output.AnchorModule, bool) {
	for _, m := range r.Modules {
		if m.Module == name {
			return m, true
		}
	}
	return output.AnchorModule{}, false
}

// Calibration ①: non-source modules and test files are dropped from anchors so
// exemplars are core source, not config/example/test.
func TestBuildAnchors_ExcludesNonSourceAndTests(t *testing.T) {
	stats := []AnchorStat{
		{Module: "svc/order", File: "svc/order/order.go", DecayedMass: 90, Lines: 100, Contributors: 4},              // core source -> keep
		{Module: "svc/order", File: "svc/order/order_test.go", DecayedMass: 80, Lines: 90, Contributors: 3},          // test file -> drop
		{Module: "svc/order", File: "svc/order/ts-tests/cov.ts", DecayedMass: 75, Lines: 85, Contributors: 3},        // test DIR -> drop
		{Module: "examples/counter", File: "examples/counter/slice.ts", DecayedMass: 70, Lines: 80, Contributors: 3}, // non-source module -> drop
		{Module: ".", File: "tsup.config.ts", DecayedMass: 60, Lines: 70, Contributors: 3},                           // root catch-all -> drop
	}

	// Excludes ON: only core source svc/order/order.go survives.
	on := buildAnchors(stats, 3, true, true, stubDigest)
	if len(on.Modules) != 1 || on.Modules[0].Module != "svc/order" {
		t.Fatalf("excludes on: modules = %+v, want only svc/order", on.Modules)
	}
	if len(on.Modules[0].Anchors) != 1 || on.Modules[0].Anchors[0].File != "svc/order/order.go" {
		t.Errorf("excludes on: test file / non-source not dropped: %+v", on.Modules[0].Anchors)
	}

	// Excludes OFF: everything contested comes back — svc/order (2 files),
	// examples/counter, and "." = 3 modules.
	off := buildAnchors(stats, 3, false, false, stubDigest)
	if len(off.Modules) != 3 {
		t.Errorf("excludes off: modules = %d, want 3", len(off.Modules))
	}
	if _, ok := moduleByName(off, "examples/counter"); !ok {
		t.Error("excludes off: examples/counter should be present")
	}
}

// Calibration ②: digest starts at real logic, not the import/comment preamble.
func TestDigestFromLines_SkipsPreamble(t *testing.T) {
	goFile := []string{
		"// Copyright header",
		"package order",
		"",
		"import (",
		"\t\"fmt\"",
		"\t\"strings\"",
		")",
		"",
		"// Order represents an order.",
		"type Order struct {",
		"\tID string",
		"}",
	}
	digest, lr := digestFromLines(goFile)
	if strings.HasPrefix(digest, "package") || strings.Contains(digest, "import (") || strings.Contains(digest, "\"fmt\"") {
		t.Errorf("digest should skip package/import preamble, got:\n%s", digest)
	}
	// First real logic is the doc comment attached to the type... comments before
	// the first logic line are preamble, so it should start at the type decl.
	if !strings.Contains(digest, "type Order struct") {
		t.Errorf("digest should include the type declaration, got:\n%s", digest)
	}
	if lr[0] != 10 { // 1-based line of "type Order struct {"
		t.Errorf("line_range start = %d, want 10 (the type decl)", lr[0])
	}

	tsFile := []string{
		"import { createSlice } from '@reduxjs/toolkit'",
		"import type { PayloadAction } from '@reduxjs/toolkit'",
		"",
		"export const counterSlice = createSlice({",
		"  name: 'counter',",
		"  initialState,",
		"})",
	}
	d2, lr2 := digestFromLines(tsFile)
	if strings.Contains(d2, "import ") {
		t.Errorf("ts digest should skip imports, got:\n%s", d2)
	}
	if !strings.HasPrefix(d2, "export const counterSlice") {
		t.Errorf("ts digest should start at the slice definition, got:\n%s", d2)
	}
	if lr2[0] != 4 {
		t.Errorf("ts line_range start = %d, want 4", lr2[0])
	}
}

// A multi-line TS named-import block closes with "} from '...'"; the skip must
// end there and land on the function, not overrun into the body.
func TestDigestFromLines_MultiLineNamedImportBlock(t *testing.T) {
	lines := []string{
		"import {",
		"  ActionCreator,",
		"  Dispatch,",
		"} from './types'",
		"",
		"/**",
		" * bind action creators.",
		" */",
		"export default function bindActionCreators(actionCreators, dispatch) {",
		"  return {}",
		"}",
	}
	d, lr := digestFromLines(lines)
	if !strings.HasPrefix(d, "export default function bindActionCreators") {
		t.Errorf("should land on the function, got:\n%s", d)
	}
	if lr[0] != 9 {
		t.Errorf("line_range start = %d, want 9 (the function)", lr[0])
	}
}

func TestDigestFromLines_AllPreamble(t *testing.T) {
	if d, lr := digestFromLines([]string{"package x", "import \"y\"", ""}); d != "" || lr != [2]int{0, 0} {
		t.Errorf("all-preamble file should yield empty digest, got %q %v", d, lr)
	}
}

func TestBuildAnchors_EmptyIsWellFormed(t *testing.T) {
	r := buildAnchors(nil, 3, false, false, stubDigest)
	var buf strings.Builder
	if err := output.EncodeAnchorsJSON(&buf, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"modules": []`) {
		t.Errorf("empty report should serialize modules: [], got %s", buf.String())
	}
}
