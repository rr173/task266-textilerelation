package service

import (
	"strings"
	"testing"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/sample"
)

func TestFreezeVersionSnapshotsCounterEvidence(t *testing.T) {
	svc := newTestService(t)
	batch, _ := svc.CreateBatch("freeze-batch", "")
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
	ver, _ := svc.CreateVersion("v-freeze", "")
	ver, _ = svc.ShareVersion(ver.ID)
	if _, err := svc.AddCounterEvidence(rel.ID, "patch", "后期补片", "补片-1"); err != nil {
		t.Fatal(err)
	}
	frozen, err := svc.FreezeVersion(ver.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(frozen.ProvenanceNote, "[patch]") {
		t.Fatalf("provenance_note missing patch evidence: %q", frozen.ProvenanceNote)
	}
}
