package service

import (
	"testing"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/sample"
)

func TestDyeConflictDowngradesToPartial(t *testing.T) {
	svc := newTestService(t)
	batch, err := svc.CreateBatch("dye-batch", "")
	if err != nil {
		t.Fatal(err)
	}
	grid := "##.#,.##.,##.#,.##."
	a, _, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "frag-a", Provenance: "site-a",
		WarpCount: 12, WeftCount: 12, WarpDensity: 18, WeftDensity: 16,
		MotifGrids: map[string]string{"diamond-a": grid},
	})
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "frag-b", Provenance: "site-b",
		WarpCount: 12, WeftCount: 12, WarpDensity: 18, WeftDensity: 16,
		MotifGrids: map[string]string{"diamond-b": grid},
	})
	if err != nil {
		t.Fatal(err)
	}
	tech := &model.TechniqueFeature{
		WeaveClass: model.WeaveTwill, Interlacing: model.RuleTwill31,
		TwillDirection: "right", Colorfastness: 4,
	}
	tA := *tech
	tA.SampleID = a.ID
	tA.DyeClass = "natural"
	tA.DyePigment = "茜草"
	tA.CarbonRatio = -18.5
	tB := *tech
	tB.SampleID = b.ID
	tB.DyeClass = "synthetic"
	tB.DyePigment = "茜草"
	tB.CarbonRatio = -28.0
	if _, err := svc.UpsertTechnique(&tA); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpsertTechnique(&tB); err != nil {
		t.Fatal(err)
	}
	motifsA, _ := svc.ListMotifs(a.ID)
	motifsB, _ := svc.ListMotifs(b.ID)
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
	if rel.Verdict != model.VerdictPartial {
		t.Fatalf("dye conflict = %s, want partial", rel.Verdict)
	}
}
