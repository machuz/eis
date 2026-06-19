package scorer

import (
	"math"
	"testing"
)

// A pool of >= 3 contributors is trusted: max-normalization stands, the largest
// value maps to 100.
func TestNormalize_TrustedPoolFullScale(t *testing.T) {
	got := Normalize(map[string]float64{"a": 10, "b": 5, "c": 1}, 3)
	if math.Abs(got["a"]-100) > 1e-9 {
		t.Errorf("a (max) = %v, want 100", got["a"])
	}
	if math.Abs(got["b"]-50) > 1e-9 {
		t.Errorf("b = %v, want 50", got["b"])
	}
}

// A two-person pool is not a real population: max-normalization still pins the
// leader at the top, so the whole pool is halved — untrusted, not erased.
func TestNormalize_SmallPoolHalved(t *testing.T) {
	got := Normalize(map[string]float64{"a": 10, "b": 4}, 2)
	if math.Abs(got["a"]-50) > 1e-9 {
		t.Errorf("a (max) in 2-pool = %v, want 50 (100 halved)", got["a"])
	}
	if math.Abs(got["b"]-20) > 1e-9 {
		t.Errorf("b in 2-pool = %v, want 20 (40 halved)", got["b"])
	}
}

// A lone author is halved, NOT zeroed: survival/design/breadth are real,
// observable facts even with no one to rank against. The relational collapse a
// solo author should see is gravity's job (Catalysis == 0), not this layer's.
func TestNormalize_SoloAuthorDampedNotErased(t *testing.T) {
	got := Normalize(map[string]float64{"solo": 7}, 1)
	if math.Abs(got["solo"]-50) > 1e-9 {
		t.Errorf("solo author = %v, want 50 (observation preserved, hedged)", got["solo"])
	}
}

func TestPoolConfidence_Steps(t *testing.T) {
	for _, c := range []struct {
		n    int
		want float64
	}{{1, 0.5}, {2, 0.5}, {3, 1.0}, {10, 1.0}} {
		if got := poolConfidence(c.n); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("poolConfidence(%d) = %v, want %v", c.n, got, c.want)
		}
	}
}
