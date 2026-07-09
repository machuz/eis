package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// AnchorsReport is the result of `eis anchors`: per module, the surviving
// exemplar code regions (propagation anchors) an agent should write toward.
//
// FIREWALL: anchors are code + weights, never authorship. contested_by_n is an
// anonymous headcount; no owner name is emitted. "Write like this surviving
// pattern", not "write like Alice".
type AnchorsReport struct {
	Modules []AnchorModule `json:"modules"`
}

type AnchorModule struct {
	Module  string   `json:"module"`
	Anchors []Anchor `json:"anchors"`
}

type Anchor struct {
	File         string  `json:"file"`
	LineRange    [2]int  `json:"line_range"`
	Survival     float64 `json:"survival"`       // mean line survival, 0-1
	Gravity      float64 `json:"gravity"`        // surviving decayed mass
	ContestedByN int     `json:"contested_by_n"` // distinct contributors (anonymous)
	Digest       string  `json:"digest"`         // compact representative excerpt
}

// EncodeAnchorsJSON writes the report as indented JSON to w.
func EncodeAnchorsJSON(w io.Writer, r AnchorsReport) error {
	if r.Modules == nil {
		r.Modules = []AnchorModule{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// PrintAnchorsTable renders a compact human view to stdout.
func PrintAnchorsTable(r AnchorsReport) {
	title := color.New(color.FgHiCyan, color.Bold)
	dim := color.New(color.FgHiBlack)
	for _, m := range r.Modules {
		fmt.Println()
		title.Printf("Anchors — %s\n", m.Module)
		for _, a := range m.Anchors {
			fmt.Printf("  %s:%d-%d\n", a.File, a.LineRange[0], a.LineRange[1])
			dim.Printf("    survival %.2f · gravity %.1f · contested by %d\n", a.Survival, a.Gravity, a.ContestedByN)
		}
	}
	if len(r.Modules) == 0 {
		dim.Fprintln(os.Stdout, "No surviving anchors found.")
	}
}
