package service

import (
	"errors"
	"testing"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/sample"
	"task266-textilerelation/internal/store"
)

func TestStaleVerdictUpdateRejected(t *testing.T) {
	svc := newTestService(t)
	batch, _ := svc.CreateBatch("cas-batch", "")
	a, _, _ := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "a", Provenance: "p1",
		WarpCount: 8, WeftCount: 8, WarpDensity: 16, WeftDensity: 14,
		MotifGrids: map[string]string{"m": "##,##"},
	})
	b, _, _ := svc.ImportSample(sample.IngestInput{
		BatchID: batch.ID, Name: "b", Provenance: "p2",
		WarpCount: 8, WeftCount: 8, WarpDensity: 16, WeftDensity: 14,
		MotifGrids: map[string]string{"m": "##,##"},
	})
	ma, _ := svc.ListMotifs(a.ID)
	mb, _ := svc.ListMotifs(b.ID)
	ma0, _ := svc.ParseMotif(ma[0].ID)
	mb0, _ := svc.ParseMotif(mb[0].ID)
	tech := &model.TechniqueFeature{
		WeaveClass: model.WeaveTwill, Interlacing: model.RuleTwill31,
		TwillDirection: "right", DyeClass: "natural", DyePigment: "靛蓝",
		Colorfastness: 4, CarbonRatio: -25.0,
	}
	tA := *tech
	tA.SampleID = a.ID
	tB := *tech
	tB.SampleID = b.ID
	_, _ = svc.UpsertTechnique(&tA)
	_, _ = svc.UpsertTechnique(&tB)
	rel, err := svc.Store.Relations.Create(&model.RelationCandidate{
		FromUnitID: ma0.ID,
		ToUnitID:   mb0.ID,
		Verdict:    model.VerdictCandidate,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Store.Relations.UpdateVerdict(rel.ID, model.VerdictCandidate, model.VerdictConfirmed, "ok"); err != nil {
		t.Fatal(err)
	}
	err = svc.Store.Relations.UpdateVerdict(rel.ID, model.VerdictCandidate, model.VerdictRejected, "stale")
	var sm *store.StateMismatchError
	if !errors.As(err, &sm) {
		t.Fatalf("stale expect should mismatch, got %v", err)
	}
}
