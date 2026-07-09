package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// GraveyardReport is the result of `eis graveyard`: per module, where past
// attempts died — code people wrote and OTHERS tore out, repeatedly, in files
// that still exist. It is the complement of anchors (surviving exemplars).
//
// FIREWALL: structure only — death_intensity / death_events / contested_by_n
// (anonymous). No author names.
type GraveyardReport struct {
	Modules []GraveyardModule `json:"modules"`
}

type GraveyardModule struct {
	Module         string             `json:"module"`
	DeathIntensity float64            `json:"death_intensity"` // others-contested dead lines / total writes, 0-1
	DeathEvents    int                `json:"death_events"`    // B≠A deletion events in repeated-death files
	Hotspots       []GraveyardHotspot `json:"hotspots"`

	// Reserved (needs an LLM; v0 omits): a semantic account of WHICH approach
	// died here and WHAT replaced it.
}

type GraveyardHotspot struct {
	File         string `json:"file"`
	Deaths       int    `json:"deaths"`
	ContestedByN int    `json:"contested_by_n"`
}

// EncodeGraveyardJSON writes the report as indented JSON to w.
func EncodeGraveyardJSON(w io.Writer, r GraveyardReport) error {
	if r.Modules == nil {
		r.Modules = []GraveyardModule{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// PrintGraveyardTable renders a compact human view to stdout.
func PrintGraveyardTable(r GraveyardReport) {
	title := color.New(color.FgHiCyan, color.Bold)
	dim := color.New(color.FgHiBlack)
	for _, m := range r.Modules {
		fmt.Println()
		title.Printf("Graveyard — %s\n", m.Module)
		dim.Printf("  death intensity %.2f · %d death events\n", m.DeathIntensity, m.DeathEvents)
		for _, h := range m.Hotspots {
			fmt.Printf("    %s  (deaths %d · contested by %d)\n", h.File, h.Deaths, h.ContestedByN)
		}
	}
	if len(r.Modules) == 0 {
		dim.Fprintln(os.Stdout, "No repeated others-contested deaths found.")
	}
}
