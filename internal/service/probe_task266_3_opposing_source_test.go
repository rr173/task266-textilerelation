package service

import (
	"testing"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/sample"
)

func TestOpposingSourceAutoRejectsCandidate(t *testing.T) {
	svc := newTestService(t)
	batch, _ := svc.CreateBatch("evidence-batch", "")
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
	rel, _ := svc.CompareMotifs(ma0.ID, mb0.ID)
	_, _ = svc.ApplyVerdict(rel.ID, model.VerdictConfirmed, "confirmed")
	if _, err := svc.AddCounterEvidence(rel.ID, "opposing_source", "年代差300年", "档案-1"); err != nil {
		t.Fatal(err)
	}
	got, _, err := svc.GetRelation(rel.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != model.VerdictRejected {
		t.Fatalf("opposing source should auto-reject, got %s", got.Verdict)
	}
}
