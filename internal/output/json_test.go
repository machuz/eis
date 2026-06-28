package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/machuz/eis/v2/internal/scorer"
)

// TestJSONWriter_EmitsLifetimeGravity guards the wiring that a unit test on the
// scorer alone cannot see: LifetimeGravity must travel from scorer.Result all
// the way into the emitted JSON, through BOTH member construction sites
// (AddDomain and AddPerRepo). An earlier pass added the struct field but forgot
// the per-site mapping, so the key serialized as a constant 0 — present but
// dead. This pins both the key and the value.
func TestJSONWriter_EmitsLifetimeGravity(t *testing.T) {
	res := []scorer.Result{{Author: "alice", Gravity: 2.7, LifetimeGravity: 10.4}}

	for _, c := range []struct {
		name string
		emit func(*JSONWriter)
	}{
		{"AddDomain", func(w *JSONWriter) { w.AddDomain("BE", 1, res, nil) }},
		{"AddPerRepo", func(w *JSONWriter) {
			w.AddDomain("BE", 1, nil, nil) // AddPerRepo attaches to an existing domain
			w.AddPerRepo("BE", "repo", res)
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := NewJSONWriter()
			c.emit(w)
			b, err := json.Marshal(w.output)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if s := string(b); !strings.Contains(s, `"lifetime_gravity":10.4`) {
				t.Fatalf("%s: lifetime_gravity not mapped from scorer.Result, got: %s", c.name, s)
			}
		})
	}
}
