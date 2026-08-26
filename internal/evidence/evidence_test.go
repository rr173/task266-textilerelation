package evidence

import (
	"testing"

	"task266-textilerelation/internal/model"
)

func TestAssessImpactRejectsOpposingSource(t *testing.T) {
	items := []*model.CounterEvidence{{Kind: "opposing_source", Description: "年代冲突"}}
	imp := AssessImpact(items, model.VerdictConfirmed)
	if !imp.Reject || imp.Suggested != model.VerdictRejected {
		t.Fatalf("opposing source should reject, got %+v", imp)
	}
}

func TestAssessImpactDowngradesPatchEvidence(t *testing.T) {
	items := []*model.CounterEvidence{{Kind: "patch", Description: "补片"}}
	imp := AssessImpact(items, model.VerdictConfirmed)
	if !imp.Downgrade || imp.Suggested != model.VerdictPartial {
		t.Fatalf("patch should downgrade, got %+v", imp)
	}
}

func TestValidateKindRejectsUnknown(t *testing.T) {
	if err := ValidateKind("unknown"); err == nil {
		t.Fatal("expected invalid kind error")
	}
}
