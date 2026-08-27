package service

import (
	"path/filepath"
	"testing"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/sample"
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

// makeConfirmedRelation 构造一对同工艺同染料的样本单元并比较出 confirmed 关系，
// 复用端到端场景的 B-C 传承路径。
func makeConfirmedRelation(t *testing.T, svc *Service) (*model.RelationCandidate, *model.RelationVersion) {
	t.Helper()
	grid := "##.#,.###,##.#,.###"
	batch, err := svc.CreateBatch("t", "")
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}
	b, _, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "f-b", Provenance: "P1",
		WarpCount: 12, WeftCount: 12, WarpDensity: 24, WeftDensity: 22,
		MotifGrids: map[string]string{"m-b": grid},
	})
	if err != nil {
		t.Fatalf("import b: %v", err)
	}
	c, _, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "f-c", Provenance: "P1",
		WarpCount: 12, WeftCount: 12, WarpDensity: 24, WeftDensity: 22,
		MotifGrids: map[string]string{"m-c": grid},
	})
	if err != nil {
		t.Fatalf("import c: %v", err)
	}
	for _, s := range []*model.FabricSample{b, c} {
		if _, err := svc.UpsertTechnique(&model.TechniqueFeature{
			SampleID: s.ID, WeaveClass: model.WeaveTwill, Interlacing: model.RuleTwill31,
			TwillDirection: "right", DyeClass: "natural", DyePigment: "靛蓝",
			Colorfastness: 4, CarbonRatio: -25.0,
		}); err != nil {
			t.Fatalf("technique sample %d: %v", s.ID, err)
		}
	}
	mb, err := svc.ListMotifs(b.ID)
	if err != nil || len(mb) != 1 {
		t.Fatalf("motifs b: %v len=%d", err, len(mb))
	}
	mc, err := svc.ListMotifs(c.ID)
	if err != nil || len(mc) != 1 {
		t.Fatalf("motifs c: %v len=%d", err, len(mc))
	}
	unitB, err := svc.ParseMotif(mb[0].ID)
	if err != nil {
		t.Fatalf("parse b: %v", err)
	}
	unitC, err := svc.ParseMotif(mc[0].ID)
	if err != nil {
		t.Fatalf("parse c: %v", err)
	}
	rel, err := svc.CompareMotifs(unitB.ID, unitC.ID)
	if err != nil {
		t.Fatalf("compare b-c: %v", err)
	}
	if rel.Verdict != model.VerdictConfirmed {
		t.Fatalf("expected confirmed, got %s", rel.Verdict)
	}
	v, err := svc.CreateVersion("v1", "")
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	if _, err := svc.ShareVersion(v.ID); err != nil {
		t.Fatalf("share version: %v", err)
	}
	return rel, v
}

// TestAddEvidenceRewritesUnlockedRelation：未入版本的关系追加补片反证，
// 应自动降级为 partial（自动改写逻辑本身仍生效）。
func TestAddEvidenceRewritesUnlockedRelation(t *testing.T) {
	svc := newTestService(t)
	grid := "##.#,.###,##.#,.###"
	batch, err := svc.CreateBatch("t", "")
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}
	b, _, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "f-b", Provenance: "P1",
		WarpCount: 12, WeftCount: 12, WarpDensity: 24, WeftDensity: 22,
		MotifGrids: map[string]string{"m-b": grid},
	})
	if err != nil {
		t.Fatalf("import b: %v", err)
	}
	c, _, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "f-c", Provenance: "P1",
		WarpCount: 12, WeftCount: 12, WarpDensity: 24, WeftDensity: 22,
		MotifGrids: map[string]string{"m-c": grid},
	})
	if err != nil {
		t.Fatalf("import c: %v", err)
	}
	for _, s := range []*model.FabricSample{b, c} {
		if _, err := svc.UpsertTechnique(&model.TechniqueFeature{
			SampleID: s.ID, WeaveClass: model.WeaveTwill, Interlacing: model.RuleTwill31,
			TwillDirection: "right", DyeClass: "natural", DyePigment: "靛蓝",
			Colorfastness: 4, CarbonRatio: -25.0,
		}); err != nil {
			t.Fatalf("technique sample %d: %v", s.ID, err)
		}
	}
	mb, _ := svc.ListMotifs(b.ID)
	mc, _ := svc.ListMotifs(c.ID)
	unitB, _ := svc.ParseMotif(mb[0].ID)
	unitC, _ := svc.ParseMotif(mc[0].ID)
	rel, err := svc.CompareMotifs(unitB.ID, unitC.ID)
	if err != nil || rel.Verdict != model.VerdictConfirmed {
		t.Fatalf("expected confirmed, got %v %s", err, rel.Verdict)
	}
	// 未入版本 → 追加补片反证应自动降级为 partial。
	if _, err := svc.AddCounterEvidence(rel.ID, "patch", "后期补片", "档案-1"); err != nil {
		t.Fatalf("add evidence: %v", err)
	}
	got, _, err := svc.GetRelation(rel.ID)
	if err != nil {
		t.Fatalf("get relation: %v", err)
	}
	if got.Verdict != model.VerdictPartial {
		t.Fatalf("unlocked relation should auto-downgrade to partial, got %s", got.Verdict)
	}
}

// TestAddEvidencePreservesVersionLockedRelation：已共享入版本的关系追加补片反证，
// 反证照常登记，但裁决不再被自动改写（confirmed 不应降级为 partial）。
func TestAddEvidencePreservesVersionLockedRelation(t *testing.T) {
	svc := newTestService(t)
	rel, _ := makeConfirmedRelation(t, svc)

	if _, err := svc.AddCounterEvidence(rel.ID, "patch", "B 单元疑为后期补片", "档案-88"); err != nil {
		t.Fatalf("add evidence: %v", err)
	}
	// 反证已登记。
	evs, err := svc.ListEvidence(rel.ID)
	if err != nil {
		t.Fatalf("list evidence: %v", err)
	}
	if len(evs) != 1 {
		t.Fatalf("evidence should be recorded, got %d", len(evs))
	}
	// 从数据库重新读取裁决：version_id != 0 时不得自动改写。
	got, _, err := svc.GetRelation(rel.ID)
	if err != nil {
		t.Fatalf("get relation: %v", err)
	}
	if got.VersionID == 0 {
		t.Fatalf("relation should be locked to a version, version_id=0")
	}
	if got.Verdict != model.VerdictConfirmed {
		t.Fatalf("version-locked relation must not be auto-rewritten: got %s, want %s",
			got.Verdict, model.VerdictConfirmed)
	}
}

