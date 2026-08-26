package store

import (
	"path/filepath"
	"testing"

	"task266-textilerelation/internal/model"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	st, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestBatchCreateAndList(t *testing.T) {
	st := openTestStore(t)
	b, err := st.Batches.Create("batch-a", "note")
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}
	list, err := st.Batches.List()
	if err != nil || len(list) != 1 || list[0].ID != b.ID {
		t.Fatalf("list batches = %+v err=%v", list, err)
	}
}

func TestRelationUniquePair(t *testing.T) {
	st := openTestStore(t)
	b, _ := st.Batches.Create("b", "")
	s1, _ := st.Samples.Create(&model.FabricSample{BatchID: b.ID, Name: "s1", WarpCount: 4, WeftCount: 4, WarpDensity: 10, WeftDensity: 10, SHA256: "h1", Status: model.MotifPending})
	s2, _ := st.Samples.Create(&model.FabricSample{BatchID: b.ID, Name: "s2", WarpCount: 4, WeftCount: 4, WarpDensity: 10, WeftDensity: 10, SHA256: "h2", Status: model.MotifPending})
	m1, _ := st.Motifs.Create(&model.MotifUnit{SampleID: s1.ID, Name: "m1", Width: 2, Height: 2, Grid: "##,##", Status: model.MotifValid})
	m2, _ := st.Motifs.Create(&model.MotifUnit{SampleID: s2.ID, Name: "m2", Width: 2, Height: 2, Grid: "##,##", Status: model.MotifValid})
	if _, err := st.Relations.Create(&model.RelationCandidate{FromUnitID: m1.ID, ToUnitID: m2.ID, Verdict: model.VerdictCandidate}); err != nil {
		t.Fatalf("create relation: %v", err)
	}
	if _, err := st.Relations.Create(&model.RelationCandidate{FromUnitID: m1.ID, ToUnitID: m2.ID, Verdict: model.VerdictCandidate}); err != model.ErrConflict {
		t.Fatalf("duplicate pair err = %v, want conflict", err)
	}
}

func TestUpdateVerdictConditional(t *testing.T) {
	st := openTestStore(t)
	b, _ := st.Batches.Create("b", "")
	s1, _ := st.Samples.Create(&model.FabricSample{BatchID: b.ID, Name: "s1", WarpCount: 4, WeftCount: 4, WarpDensity: 10, WeftDensity: 10, SHA256: "h1", Status: model.MotifPending})
	s2, _ := st.Samples.Create(&model.FabricSample{BatchID: b.ID, Name: "s2", WarpCount: 4, WeftCount: 4, WarpDensity: 10, WeftDensity: 10, SHA256: "h2", Status: model.MotifPending})
	m1, _ := st.Motifs.Create(&model.MotifUnit{SampleID: s1.ID, Name: "m1", Width: 2, Height: 2, Grid: "##,##", Status: model.MotifValid})
	m2, _ := st.Motifs.Create(&model.MotifUnit{SampleID: s2.ID, Name: "m2", Width: 2, Height: 2, Grid: "##,##", Status: model.MotifValid})
	rel, _ := st.Relations.Create(&model.RelationCandidate{FromUnitID: m1.ID, ToUnitID: m2.ID, Verdict: model.VerdictCandidate})
	if err := st.Relations.UpdateVerdict(rel.ID, model.VerdictCandidate, model.VerdictConfirmed, "ok"); err != nil {
		t.Fatalf("update verdict: %v", err)
	}
	if err := st.Relations.UpdateVerdict(rel.ID, model.VerdictCandidate, model.VerdictRejected, "stale"); err == nil {
		t.Fatal("expected stale expect mismatch")
	}
}
