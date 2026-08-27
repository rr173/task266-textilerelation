package service

import (
	"testing"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/sample"
)

// setupSharedVersion 复用冒烟测试的极小场景：建立一条 confirmed 关系，
// 收录进版本草稿并共享，返回服务、关系与版本。
func setupSharedVersion(t *testing.T) (*Service, *model.RelationCandidate, *model.RelationVersion) {
	t.Helper()
	svc := newTestService(t)

	batch, err := svc.CreateBatch("b", "n")
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}
	grid := "##.#,.###,##.#,.###"
	sB, _, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "frag-b", Provenance: "p",
		WarpCount: 12, WeftCount: 12, WarpDensity: 24, WeftDensity: 22,
		MotifGrids: map[string]string{"m-b": grid},
	})
	if err != nil {
		t.Fatalf("import b: %v", err)
	}
	sC, _, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "frag-c", Provenance: "p",
		WarpCount: 12, WeftCount: 12, WarpDensity: 24, WeftDensity: 22,
		MotifGrids: map[string]string{"m-c": grid},
	})
	if err != nil {
		t.Fatalf("import c: %v", err)
	}
	for _, sid := range []int64{sB.ID, sC.ID} {
		if _, err := svc.UpsertTechnique(&model.TechniqueFeature{
			SampleID: sid, WeaveClass: model.WeaveTwill, Interlacing: model.RuleTwill31,
			TwillDirection: "right", DyeClass: "natural", DyePigment: "靛蓝",
			Colorfastness: 4, CarbonRatio: -25.0,
		}); err != nil {
			t.Fatalf("technique %d: %v", sid, err)
		}
	}
	mB, err := svc.ListMotifs(sB.ID)
	if err != nil || len(mB) != 1 {
		t.Fatalf("motifs b: %v len=%d", err, len(mB))
	}
	mC, err := svc.ListMotifs(sC.ID)
	if err != nil || len(mC) != 1 {
		t.Fatalf("motifs c: %v len=%d", err, len(mC))
	}
	unitB, err := svc.ParseMotif(mB[0].ID)
	if err != nil {
		t.Fatalf("parse b: %v", err)
	}
	unitC, err := svc.ParseMotif(mC[0].ID)
	if err != nil {
		t.Fatalf("parse c: %v", err)
	}
	rel, err := svc.CompareMotifs(unitB.ID, unitC.ID)
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if rel.Verdict != model.VerdictConfirmed {
		t.Fatalf("expected confirmed, got %s", rel.Verdict)
	}
	v, err := svc.CreateVersion("v", "")
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	if _, err := svc.ShareVersion(v.ID); err != nil {
		t.Fatalf("share version: %v", err)
	}
	return svc, rel, v
}

// TestShareVersionLocksRelationVersionID 验证共享版本时收录关系的 version_id 被写入，
// 使后续追加反证时版本锁定生效（不再自动改写裁决）。
func TestShareVersionLocksRelationVersionID(t *testing.T) {
	svc, rel, v := setupSharedVersion(t)

	got, _, err := svc.GetRelation(rel.ID)
	if err != nil {
		t.Fatalf("get relation: %v", err)
	}
	if got.VersionID != v.ID {
		t.Fatalf("relation version_id not locked: got %d want %d", got.VersionID, v.ID)
	}
}

// TestSharedRelationNotAutoRewrittenByCounterEvidence 验证已入版本的关系在追加
// 补片反证时裁决维持不变（版本锁定生效）。
func TestSharedRelationNotAutoRewrittenByCounterEvidence(t *testing.T) {
	svc, rel, _ := setupSharedVersion(t)

	if _, err := svc.AddCounterEvidence(rel.ID, "patch", "疑为后期补片", "档案-88"); err != nil {
		t.Fatalf("add evidence: %v", err)
	}
	got, _, err := svc.GetRelation(rel.ID)
	if err != nil {
		t.Fatalf("get relation: %v", err)
	}
	if got.Verdict != model.VerdictConfirmed {
		t.Fatalf("shared relation verdict auto-rewritten: got %s want %s",
			got.Verdict, model.VerdictConfirmed)
	}
}
