package scorer

import "testing"

// classifyRoleName is a tiny helper so the table tests read as role assertions.
func classifyRoleName(r Result) string {
	return classifyRole(r).Name
}

// High Catalysis + high Production used to classify as Producer, because
// Producer's notLow(Production)=1.0 beat Anchor's highness(Catalysis)=0.70 by
// more than the precedence margin. With Producer capped at producerCeiling
// (0.65), a Catalysis that reaches the "high" band wins the tiebreak → Anchor.
// (A real high-output, high-catalysis member: Catalysis 68, Production 100.)
func TestClassifyRole_HighCatalysisHighProductionIsAnchorNotProducer(t *testing.T) {
	r := Result{Catalysis: 68, Production: 100, Survival: 20, Design: 0, Breadth: 100, DebtCleanup: 71}
	if got := classifyRoleName(r); got != "Anchor" {
		t.Fatalf("Catalysis 68 + Production 100: want Anchor, got %q", got)
	}
}

// Catalysis at the start of the high band (60 → highness 0.5) tips the role to
// Anchor over a maxed Producer. producerCeiling=0.64 keeps this boundary
// inclusive (0.64-0.5=0.14 <= priorityMargin 0.15) without a float edge.
func TestClassifyRole_CatalysisInHighBandIsAnchor(t *testing.T) {
	r := Result{Catalysis: 60, Production: 100}
	if got := classifyRoleName(r); got != "Anchor" {
		t.Fatalf("Catalysis 60 + Production 100: want Anchor, got %q", got)
	}
}

// Production high but Catalysis below the high band stays Producer — the cap
// must not turn every high-output member into an Anchor.
func TestClassifyRole_HighProductionLowCatalysisStaysProducer(t *testing.T) {
	r := Result{Catalysis: 40, Production: 100, Survival: 20, Design: 10, Breadth: 40, DebtCleanup: 10}
	if got := classifyRoleName(r); got != "Producer" {
		t.Fatalf("Catalysis 40 + Production 100: want Producer, got %q", got)
	}
}

// A pure high-output member with nothing else is still a Producer (the capped
// score, 0.65, clears minConf and beats absent roles).
func TestClassifyRole_PureProducer(t *testing.T) {
	r := Result{Production: 90}
	if got := classifyRoleName(r); got != "Producer" {
		t.Fatalf("Production 90 only: want Producer, got %q", got)
	}
}
