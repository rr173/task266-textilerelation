package service

import (
	"errors"
	"testing"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/sample"
)

func TestCreateVersionRejectsMutualCycle(t *testing.T) {
	svc := newTestService(t)
	batch, _ := svc.CreateBatch("cycle-batch", "")
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
	rAB, _ := svc.CompareMotifs(ma0.ID, mb0.ID)
	rBA, _ := svc.CompareMotifs(mb0.ID, ma0.ID)
	_, _ = svc.ApplyVerdict(rAB.ID, model.VerdictConfirmed, "ab")
	_, _ = svc.ApplyVerdict(rBA.ID, model.VerdictConfirmed, "ba")
	if _, err := svc.CreateVersion("v1", ""); !errors.Is(err, model.ErrCycle) {
		t.Fatalf("mutual confirmed cycle should reject, got %v", err)
	}
}
