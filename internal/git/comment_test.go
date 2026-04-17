package git

import "testing"

func TestDetectLang(t *testing.T) {
	cases := []struct {
		name string
		path string
		want LangKind
	}{
		{"go", "internal/foo.go", LangCStyle},
		{"ts", "src/app.ts", LangCStyle},
		{"tsx", "src/Page.tsx", LangCStyle},
		{"python", "scripts/build.py", LangPython},
		{"ruby", "lib/foo.rb", LangRuby},
		{"shell", "scripts/deploy.sh", LangHashStyle},
		{"yaml", ".github/workflows/ci.yml", LangHashStyle},
		{"dockerfile", "Dockerfile", LangHashStyle},
		{"html", "templates/index.html", LangHTML},
		{"sql", "migrations/001.sql", LangSQL},
		{"markdown", "README.md", LangProse},
		{"text", "notes.txt", LangProse},
		{"unknown", "vendor/data.bin", LangUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := DetectLang(c.path); got != c.want {
				t.Errorf("DetectLang(%q) = %v, want %v", c.path, got, c.want)
			}
		})
	}
}

func TestNewFileFilterNilForProse(t *testing.T) {
	if NewFileFilter("README.md") != nil {
		t.Error("prose file should return nil filter")
	}
	if NewFileFilter("vendor/data.bin") != nil {
		t.Error("unknown file should return nil filter")
	}
	if NewFileFilter("main.go") == nil {
		t.Error("code file should return a filter")
	}
}

func TestNilFilterKeepsAllLines(t *testing.T) {
	var f *FileFilter
	for _, line := range []string{"", "   ", "// comment", "real code"} {
		if f.IsSkip(line) {
			t.Errorf("nil filter must not skip %q", line)
		}
	}
}

type lineCase struct {
	line string
	skip bool
}

func runLineCases(t *testing.T, filename string, cases []lineCase) {
	t.Helper()
	f := NewFileFilter(filename)
	for i, c := range cases {
		got := f.IsSkip(c.line)
		if got != c.skip {
			t.Errorf("case %d: IsSkip(%q) = %v, want %v", i, c.line, got, c.skip)
		}
	}
}

func TestGoFilter(t *testing.T) {
	runLineCases(t, "main.go", []lineCase{
		{"package main", false},
		{"", true},
		{"// single comment", true},
		{"    // indented comment", true},
		{"x := 1 // trailing comment", false}, // trailing comments count as code
		{"/* short */", true},
		{"/* open", true},
		{" * continuation", true}, // inside block
		{" */", true},              // block end
		{"func Foo() {}", false},
		{"/* wrapped */code()", false}, // block ends with code after
	})
}

func TestGoJSDocBlock(t *testing.T) {
	f := NewFileFilter("lib.go")
	lines := []lineCase{
		{"/**", true},
		{" * Multi-line", true},
		{" * doc comment.", true},
		{" */", true},
		{"func X() {}", false},
	}
	for i, c := range lines {
		if got := f.IsSkip(c.line); got != c.skip {
			t.Errorf("case %d: IsSkip(%q) = %v, want %v", i, c.line, got, c.skip)
		}
	}
}

func TestPythonFilter(t *testing.T) {
	runLineCases(t, "foo.py", []lineCase{
		{"import os", false},
		{"# hash comment", true},
		{`"""docstring"""`, true},
		{`"""start`, true},
		{"middle of docstring", true},
		{`"""`, true}, // end
		{"def foo():", false},
		{"    pass", false},
	})
}

func TestPythonSingleLineDocstring(t *testing.T) {
	runLineCases(t, "foo.py", []lineCase{
		{`x = """hi"""`, false}, // assignment with triple-quote string — NOT a docstring start
		{`"""hi"""`, true},       // standalone triple-quote string — treat as docstring
	})
}

func TestShellFilter(t *testing.T) {
	runLineCases(t, "deploy.sh", []lineCase{
		{"#!/bin/bash", true},
		{"# comment", true},
		{"echo hi", false},
		{"", true},
	})
}

func TestYAMLFilter(t *testing.T) {
	runLineCases(t, "ci.yml", []lineCase{
		{"# yaml comment", true},
		{"name: build", false},
		{"  - run: make", false},
	})
}

func TestHTMLFilter(t *testing.T) {
	runLineCases(t, "page.html", []lineCase{
		{"<!-- inline -->", true},
		{"<!-- start", true},
		{"still in comment", true},
		{"end -->", true},
		{"<div>hi</div>", false},
	})
}

func TestSQLFilter(t *testing.T) {
	runLineCases(t, "m.sql", []lineCase{
		{"-- sql comment", true},
		{"/* block */", true},
		{"SELECT 1;", false},
	})
}

func TestProseFileUnfiltered(t *testing.T) {
	f := NewFileFilter("paper.md")
	if f != nil {
		t.Fatal("prose filter should be nil")
	}
	// blank lines and comment-looking text in prose must count.
	for _, line := range []string{"", "// looks-like-code", "# heading"} {
		if f.IsSkip(line) {
			t.Errorf("prose filter must not skip %q", line)
		}
	}
}

func TestBlockReopens(t *testing.T) {
	// Verify that after one block closes, a subsequent block can open correctly.
	f := NewFileFilter("a.go")
	seq := []lineCase{
		{"/* first", true},
		{"inside", true},
		{"*/", true},
		{"code()", false},
		{"/* second", true},
		{"still inside", true},
		{"*/", true},
		{"more_code()", false},
	}
	for i, c := range seq {
		if got := f.IsSkip(c.line); got != c.skip {
			t.Errorf("step %d: IsSkip(%q) = %v, want %v", i, c.line, got, c.skip)
		}
	}
}
