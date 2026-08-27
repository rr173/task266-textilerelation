package service

import (
	"path/filepath"
	"testing"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/store"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	path := filepath.Join(t.TempDir(), "svc.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return New(st)
}

func TestSelfCheckOpensDatabase(t *testing.T) {
	svc := newTestService(t)
	out, err := svc.SelfCheck()
	if err != nil {
		t.Fatalf("self check: %v", err)
	}
	if out["db"] != "ok" {
		t.Fatalf("db status = %q", out["db"])
	}
}

// makeConfirmedRelation 在测试库中直接构造一条 confirmed 关系，
// 模拟"已复核确认"的传承关系，用于验证反证驱动的自动裁决链路。
func makeConfirmedRelation(t *testing.T, svc *Service) *model.RelationCandidate {
	t.Helper()
	b, err := svc.Store.Batches.Create("b", "")
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}
	s1, err := svc.Store.Samples.Create(&model.FabricSample{
		BatchID: b.ID, Name: "s1", WarpCount: 4, WeftCount: 4,
		WarpDensity: 10, WeftDensity: 10, SHA256: "h1", Status: model.MotifPending,
	})
	if err != nil {
		t.Fatalf("create sample s1: %v", err)
	}
	s2, err := svc.Store.Samples.Create(&model.FabricSample{
		BatchID: b.ID, Name: "s2", WarpCount: 4, WeftCount: 4,
		WarpDensity: 10, WeftDensity: 10, SHA256: "h2", Status: model.MotifPending,
	})
	if err != nil {
		t.Fatalf("create sample s2: %v", err)
	}
	m1, err := svc.Store.Motifs.Create(&model.MotifUnit{
		SampleID: s1.ID, Name: "m1", Width: 2, Height: 2,
		Grid: "##,##", Status: model.MotifValid,
	})
	if err != nil {
		t.Fatalf("create motif m1: %v", err)
	}
	m2, err := svc.Store.Motifs.Create(&model.MotifUnit{
		SampleID: s2.ID, Name: "m2", Width: 2, Height: 2,
		Grid: "##,##", Status: model.MotifValid,
	})
	if err != nil {
		t.Fatalf("create motif m2: %v", err)
	}
	rel, err := svc.Store.Relations.Create(&model.RelationCandidate{
		FromUnitID: m1.ID, ToUnitID: m2.ID,
		Verdict:    model.VerdictConfirmed,
		Summary:    "人工复核确认",
	})
	if err != nil {
		t.Fatalf("create relation: %v", err)
	}
	return rel
}

// TestAddCounterEvidenceOpposingSourceAutoRejects 验证出处年代冲突反证应触发
// 自动否决：已确认的传承关系在追加 opposing_source 反证后裁决转为 rejected。
//
// 该用例回归一个缺陷——AssessImpact 漏判 opposing_source 类型，导致反证既不
// 置 Reject 也不置 Downgrade，Suggested 维持当前裁决，自动裁决链路不触发，
// 已确认关系在出处冲突反证下仍保持 confirmed。
func TestAddCounterEvidenceOpposingSourceAutoRejects(t *testing.T) {
	svc := newTestService(t)
	rel := makeConfirmedRelation(t, svc)

	if _, err := svc.AddCounterEvidence(rel.ID, "opposing_source", "出处年代冲突", "档案-401"); err != nil {
		t.Fatalf("add evidence: %v", err)
	}
	updated, err := svc.Store.Relations.Get(rel.ID)
	if err != nil {
		t.Fatalf("get relation: %v", err)
	}
	if updated.Verdict != model.VerdictRejected {
		t.Fatalf("opposing_source should auto-reject confirmed relation, got verdict=%s", updated.Verdict)
	}
}

// TestAddCounterEvidencePatchAutoDowngrades 验证补片反证触发降级复核，
// 与 opposing_source 的否决路径区分，确保两类反证各自走对应自动裁决。
func TestAddCounterEvidencePatchAutoDowngrades(t *testing.T) {
	svc := newTestService(t)
	rel := makeConfirmedRelation(t, svc)

	if _, err := svc.AddCounterEvidence(rel.ID, "patch", "纹样疑为后期补片", "档案-88"); err != nil {
		t.Fatalf("add evidence: %v", err)
	}
	updated, err := svc.Store.Relations.Get(rel.ID)
	if err != nil {
		t.Fatalf("get relation: %v", err)
	}
	if updated.Verdict != model.VerdictPartial {
		t.Fatalf("patch should downgrade confirmed relation to partial, got verdict=%s", updated.Verdict)
	}
}
