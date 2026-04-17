package git

import (
	"path/filepath"
	"strings"
)

// LangKind identifies how a file should be scanned for comment lines.
type LangKind int

const (
	// LangUnknown files pass through unfiltered (unknown extensions).
	LangUnknown LangKind = iota
	// LangProse (.md / .txt / .rst) — recorded verbatim for papers and prose analysis.
	LangProse
	// LangCStyle — //, /* */
	LangCStyle
	// LangHashStyle — # only (YAML, Shell, TOML, Dockerfile)
	LangHashStyle
	// LangPython — #, """ ... """, ''' ... '''
	LangPython
	// LangRuby — #, =begin/=end
	LangRuby
	// LangHTML — <!-- -->
	LangHTML
	// LangSQL — --, /* */
	LangSQL
	// LangLua — --, --[[ ]]
	LangLua
	// LangHaskell — --, {- -}
	LangHaskell
)

// DetectLang returns the language kind for a given file path.
func DetectLang(filename string) LangKind {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go",
		".java",
		".c", ".cc", ".cpp", ".cxx", ".h", ".hh", ".hpp", ".hxx",
		".cs",
		".js", ".jsx", ".mjs", ".cjs",
		".ts", ".tsx",
		".rs",
		".swift",
		".kt", ".kts",
		".scala", ".sc",
		".dart",
		".php",
		".m", ".mm",
		".groovy", ".gradle":
		return LangCStyle
	case ".py", ".pyw", ".pyi":
		return LangPython
	case ".rb", ".rbi", ".rake":
		return LangRuby
	case ".sh", ".bash", ".zsh", ".fish", ".ksh",
		".yaml", ".yml",
		".toml", ".ini", ".cfg", ".conf",
		".tf", ".tfvars", ".hcl",
		".pl", ".pm", ".r":
		return LangHashStyle
	case ".html", ".htm", ".xml", ".vue", ".svelte", ".xhtml":
		return LangHTML
	case ".sql":
		return LangSQL
	case ".lua":
		return LangLua
	case ".hs", ".lhs", ".elm":
		return LangHaskell
	case ".md", ".markdown", ".mdx", ".txt", ".rst", ".adoc", ".asciidoc", ".tex", ".org":
		return LangProse
	}
	// Extensionless files — check canonical filenames.
	base := strings.ToLower(filepath.Base(filename))
	switch base {
	case "dockerfile", "makefile", "gemfile", "rakefile", "brewfile", "procfile", "vagrantfile":
		return LangHashStyle
	}
	return LangUnknown
}

// FileFilter decides which lines of a file should count toward impact metrics.
// Single-line comments, block-comment contents, and blank lines are skipped for code files.
// Prose and unknown files pass every line through unchanged.
//
// FileFilter is stateful (tracks multi-line block comments). Create one per
// logical scan (one file, one diff) and call IsSkip in line order.
// Not safe for concurrent use.
type FileFilter struct {
	lang     LangKind
	inBlock  bool
	blockEnd string
}

// NewFileFilter returns a filter for the given filename. Returns nil when the
// file is prose or unknown — callers MUST nil-check and treat nil as "keep all".
func NewFileFilter(filename string) *FileFilter {
	lang := DetectLang(filename)
	if lang == LangUnknown || lang == LangProse {
		return nil
	}
	return &FileFilter{lang: lang}
}

// IsSkip reports whether line is a comment-only or blank line and should be
// excluded from production / survival / design counts.
//
// Nil receiver is valid and always returns false (no filtering).
func (f *FileFilter) IsSkip(line string) bool {
	if f == nil {
		return false
	}
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}
	if f.inBlock {
		idx := strings.Index(trimmed, f.blockEnd)
		if idx < 0 {
			return true
		}
		// Block ends on this line. Check whatever follows the closing marker.
		f.inBlock = false
		after := strings.TrimSpace(trimmed[idx+len(f.blockEnd):])
		if after == "" {
			return true
		}
		return f.checkSingle(after)
	}
	return f.checkSingle(trimmed)
}

func (f *FileFilter) checkSingle(trimmed string) bool {
	switch f.lang {
	case LangCStyle:
		if strings.HasPrefix(trimmed, "//") {
			return true
		}
		// JSDoc/Javadoc/Doxygen continuation lines: "*", "* text", "*/".
		// We hit these when a diff hunk starts mid-block — real code
		// rarely begins a line with "*" in C-style languages.
		if trimmed == "*" || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "*/") {
			return true
		}
		return f.checkBlockStart(trimmed, "/*", "*/")
	case LangHashStyle:
		return strings.HasPrefix(trimmed, "#")
	case LangPython:
		if strings.HasPrefix(trimmed, "#") {
			return true
		}
		if strings.HasPrefix(trimmed, `"""`) {
			return f.checkBlockStart(trimmed, `"""`, `"""`)
		}
		if strings.HasPrefix(trimmed, `'''`) {
			return f.checkBlockStart(trimmed, `'''`, `'''`)
		}
		return false
	case LangRuby:
		if strings.HasPrefix(trimmed, "#") {
			return true
		}
		if strings.HasPrefix(trimmed, "=begin") {
			f.inBlock = true
			f.blockEnd = "=end"
			return true
		}
		return false
	case LangHTML:
		return f.checkBlockStart(trimmed, "<!--", "-->")
	case LangSQL:
		if strings.HasPrefix(trimmed, "--") {
			return true
		}
		return f.checkBlockStart(trimmed, "/*", "*/")
	case LangLua:
		if strings.HasPrefix(trimmed, "--[[") {
			rest := trimmed[len("--[["):]
			if idx := strings.Index(rest, "]]"); idx >= 0 {
				after := strings.TrimSpace(rest[idx+2:])
				return after == "" || f.checkSingle(after)
			}
			f.inBlock = true
			f.blockEnd = "]]"
			return true
		}
		return strings.HasPrefix(trimmed, "--")
	case LangHaskell:
		if strings.HasPrefix(trimmed, "--") {
			return true
		}
		return f.checkBlockStart(trimmed, "{-", "-}")
	}
	return false
}

func (f *FileFilter) checkBlockStart(trimmed, start, end string) bool {
	if !strings.HasPrefix(trimmed, start) {
		return false
	}
	rest := trimmed[len(start):]
	if idx := strings.Index(rest, end); idx >= 0 {
		after := strings.TrimSpace(rest[idx+len(end):])
		// "/* c */ // d" is still pure comment — recurse on whatever follows the close.
		return after == "" || f.checkSingle(after)
	}
	f.inBlock = true
	f.blockEnd = end
	return true
}
