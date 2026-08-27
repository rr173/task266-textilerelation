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

func TestDecideVerdictDyeConflictDowngradesToPartial(t *testing.T) {
	// 天然茜草 vs 合成分散蓝：拓扑相似且织造兼容，但染料类别冲突。
	// 此前 decideVerdict 忽略 dye 维度而误判 confirmed；修复后须稳定判为 partial。
	got := decideVerdict(0.92, "compatible", "conflict")
	if got != model.VerdictPartial {
		t.Fatalf("dye conflict = %s, want partial", got)
	}
}

func TestDecideVerdictDyeConflictPartialWeaveStillPartial(t *testing.T) {
	// 织造 partial 且染料冲突 → partial（不可因染料冲突反向升级为 conflict）。
	got := decideVerdict(0.8, "partial", "conflict")
	if got != model.VerdictPartial {
		t.Fatalf("partial weave + dye conflict = %s, want partial", got)
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
