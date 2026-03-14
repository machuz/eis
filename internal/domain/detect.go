package domain

import (
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// Domain represents a scoring domain (BE, FE, Infra, Firmware, or custom)
type Domain string

const (
	Backend  Domain = "Backend"
	Frontend Domain = "Frontend"
	Infra    Domain = "Infra"
	Firmware Domain = "Firmware"
	Unknown  Domain = "Unknown"
)

// builtInOrder defines display priority for built-in domains.
var builtInOrder = map[Domain]int{
	Backend: 0, Frontend: 1, Infra: 2, Firmware: 3,
}

// DefaultExtDomain is the default file extension to domain mapping.
var defaultExtDomain = map[string]Domain{
	// Backend
	".go":    Backend,
	".py":    Backend,
	".java":  Backend,
	".rb":    Backend,
	".rs":    Backend,
	".scala": Backend,
	".hs":    Backend,
	".ml":    Backend,
	".mli":   Backend,
	".php":   Backend,
	".cs":    Backend,
	".kt":    Backend,

	// Frontend
	".ts":     Frontend,
	".tsx":    Frontend,
	".js":     Frontend,
	".jsx":    Frontend,
	".vue":    Frontend,
	".svelte": Frontend,
	".css":    Frontend,
	".scss":   Frontend,
	".less":   Frontend,
	".html":   Frontend,

	// Infra
	".tf":   Infra,
	".hcl":  Infra,
	".yaml": Infra,
	".yml":  Infra,
	".toml": Infra,

	// Firmware
	".c":   Firmware,
	".h":   Firmware,
	".cpp": Firmware,
	".hpp": Firmware,
	".cc":  Firmware,
	".S":   Firmware,
	".s":   Firmware,
	".ld":  Firmware,
}

// NormalizeName converts a domain name to Title Case (e.g. "backend" → "Backend", "mobile" → "Mobile").
func NormalizeName(name string) Domain {
	if len(name) == 0 {
		return Unknown
	}
	r := []rune(name)
	r[0] = unicode.ToUpper(r[0])
	return Domain(string(r))
}

// BuildExtMap creates an extension-to-domain map by merging defaults with custom config.
// customExts maps domain names (as in config, e.g. "backend", "mobile") to extension lists.
// Custom extensions override defaults (e.g. moving .py from Backend to "Data").
// If includeDefaults is false, only custom extensions are used (no built-in mappings).
func BuildExtMap(customExts map[string][]string, includeDefaults bool) map[string]Domain {
	m := make(map[string]Domain, len(defaultExtDomain))
	if includeDefaults {
		for ext, d := range defaultExtDomain {
			m[ext] = d
		}
	}
	for name, exts := range customExts {
		d := NormalizeName(name)
		for _, ext := range exts {
			ext = strings.ToLower(ext)
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			m[ext] = d
		}
	}
	return m
}

// SortDomains returns domains in stable display order:
// built-in domains first (Backend, Frontend, Infra, Firmware), then custom alphabetically, then Unknown last.
func SortDomains(ds []Domain) []Domain {
	result := make([]Domain, len(ds))
	copy(result, ds)
	sort.SliceStable(result, func(i, j int) bool {
		di, dj := result[i], result[j]
		// Unknown always last
		if di == Unknown || dj == Unknown {
			return dj == Unknown && di != Unknown
		}
		pi, oki := builtInOrder[di]
		pj, okj := builtInOrder[dj]
		// Both built-in: compare by priority
		if oki && okj {
			return pi < pj
		}
		// Built-in before custom
		if oki != okj {
			return oki
		}
		// Both custom: alphabetical
		return di < dj
	})
	return result
}

// DetectFromFiles determines the domain of a repo based on file extension distribution.
// If extMap is nil, uses the default extension mapping.
func DetectFromFiles(files []string, extMap map[string]Domain) Domain {
	if extMap == nil {
		extMap = defaultExtDomain
	}

	counts := make(map[Domain]int)

	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		if d, ok := extMap[ext]; ok {
			counts[d]++
		}
	}

	if len(counts) == 0 {
		return Unknown
	}

	// Special case: .yaml/.yml files alone don't make it Infra.
	// Only classify as Infra if there are .tf/.hcl files present.
	if counts[Infra] > 0 {
		hasTerraform := false
		for _, f := range files {
			ext := strings.ToLower(filepath.Ext(f))
			if ext == ".tf" || ext == ".hcl" {
				hasTerraform = true
				break
			}
		}
		if !hasTerraform {
			// YAML/TOML only — likely config files, not an infra repo
			delete(counts, Infra)
		}
	}

	// Find domain with highest count (tie-break: built-in priority, then alphabetical)
	var best Domain
	bestCount := 0
	for d, c := range counts {
		if c > bestCount || (c == bestCount && betterDomain(d, best)) {
			best = d
			bestCount = c
		}
	}

	return best
}

// betterDomain returns true if a should take priority over b in a tie.
// Built-in domains have priority over custom, then alphabetical.
func betterDomain(a, b Domain) bool {
	pa, oka := builtInOrder[a]
	pb, okb := builtInOrder[b]
	if oka && okb {
		return pa < pb
	}
	if oka != okb {
		return oka
	}
	return a < b
}

// MatchRepoPattern checks if a repo name matches a glob pattern
func MatchRepoPattern(repoName string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, _ := filepath.Match(pattern, repoName)
		if matched {
			return true
		}
	}
	return false
}
