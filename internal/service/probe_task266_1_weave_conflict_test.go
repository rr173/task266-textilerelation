package service

import (
	"testing"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/sample"
)

func TestWeaveConflictMarksVisualCoincidence(t *testing.T) {
	svc := newTestService(t)
	batch, err := svc.CreateBatch("probe-batch", "")
	if err != nil {
		t.Fatal(err)
	}
	a, _, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "frag-a", Provenance: "site-a",
		WarpCount: 12, WeftCount: 12, WarpDensity: 18, WeftDensity: 16,
		MotifGrids: map[string]string{"diamond-a": "##.#,.##.,##.#,.##."},
	})
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "frag-b", Provenance: "site-b",
		WarpCount: 12, WeftCount: 12, WarpDensity: 18, WeftDensity: 16,
		MotifGrids: map[string]string{"diamond-b": "##.#,.###,##.#,.###"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpsertTechnique(&model.TechniqueFeature{
		SampleID: a.ID, WeaveClass: model.WeavePlain, Interlacing: model.RulePlain11,
		DyeClass: "natural", DyePigment: "茜草", Colorfastness: 4, CarbonRatio: -18.5,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpsertTechnique(&model.TechniqueFeature{
		SampleID: b.ID, WeaveClass: model.WeaveTwill, Interlacing: model.RuleTwill31,
		TwillDirection: "right", DyeClass: "natural", DyePigment: "靛蓝",
		Colorfastness: 4, CarbonRatio: -25.0,
	}); err != nil {
		t.Fatal(err)
	}
	motifsA, err := svc.ListMotifs(a.ID)
	if err != nil || len(motifsA) == 0 {
		t.Fatalf("motifs a: %v", err)
	}
	motifsB, err := svc.ListMotifs(b.ID)
	if err != nil || len(motifsB) == 0 {
		t.Fatalf("motifs b: %v", err)
	}
	ma, err := svc.ParseMotif(motifsA[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	mb, err := svc.ParseMotif(motifsB[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	rel, err := svc.CompareMotifs(ma.ID, mb.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rel.Verdict != model.VerdictConflict {
		t.Fatalf("plain vs twill high sim = %s, want conflict", rel.Verdict)
	}
}
