package relation

import (
	"testing"

	"task266-textilerelation/internal/model"
)

func TestDecideVerdictConflictOnWeaveMismatch(t *testing.T) {
	got := decideVerdict(0.91, "conflict", "compatible")
	if got != model.VerdictConflict {
		t.Fatalf("high sim + weave conflict = %s, want conflict", got)
	}
}

func TestDecideVerdictConfirmedOnCompatibleWeave(t *testing.T) {
	got := decideVerdict(0.85, "compatible", "compatible")
	if got != model.VerdictConfirmed {
		t.Fatalf("compatible weave = %s, want confirmed", got)
	}
}

func TestCanTransitionRejectsIllegalVerdict(t *testing.T) {
	if CanTransition(model.VerdictRejected, model.VerdictConfirmed) {
		t.Fatal("rejected -> confirmed should be illegal")
	}
}

func TestApplyVerdictUpdatesSummaryDefault(t *testing.T) {
	rel := &model.RelationCandidate{Verdict: model.VerdictCandidate}
	if err := ApplyVerdict(rel, model.VerdictConfirmed, ""); err != nil {
		t.Fatalf("apply verdict: %v", err)
	}
}
