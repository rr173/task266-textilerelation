package sample

import (
	"testing"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/store"
)

func openStores(t *testing.T) (*store.SampleStore, *store.MotifStore, *store.BatchStore) {
	t.Helper()
	st, err := store.Open(t.TempDir() + "/sample.db")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	b, err := st.Batches.Create("batch", "")
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	_ = b
	return st.Samples, st.Motifs, st.Batches
}

func TestIngestIdempotentByHash(t *testing.T) {
	samples, motifs, batches := openStores(t)
	b, _ := batches.Create("b2", "")
	in := IngestInput{
		BatchID: b.ID, Name: "frag-a", Provenance: "site-a",
		WarpCount: 8, WeftCount: 8, WarpDensity: 16, WeftDensity: 14,
		MotifGrids: map[string]string{"diamond": "##.#,.##."},
	}
	first, created, err := Ingest(samples, motifs, in)
	if err != nil || !created {
		t.Fatalf("first ingest: created=%v err=%v", created, err)
	}
	second, created2, err := Ingest(samples, motifs, in)
	if err != nil || created2 || second.ID != first.ID {
		t.Fatalf("second ingest: id=%d/%d created=%v err=%v", second.ID, first.ID, created2, err)
	}
}

func TestNextBatchStatusRequiresSamples(t *testing.T) {
	if _, err := NextBatchStatus(model.BatchPending, 0, 0); err == nil {
		t.Fatal("pending->ready without samples should fail")
	}
	if next, err := NextBatchStatus(model.BatchPending, 1, 0); err != nil || next != model.BatchReady {
		t.Fatalf("pending->ready = %s err=%v", next, err)
	}
}

func TestValidMotifsOfSample(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/motif.db")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	b, err := st.Batches.Create("b3", "")
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	smp, err := st.Samples.Create(&model.FabricSample{
		BatchID: b.ID, Name: "s", WarpCount: 4, WeftCount: 4,
		WarpDensity: 10, WeftDensity: 10, SHA256: "abc", Status: model.MotifPending,
	})
	if err != nil {
		t.Fatalf("sample: %v", err)
	}
	if _, err := st.Motifs.Create(&model.MotifUnit{
		SampleID: smp.ID, Name: "m", Width: 2, Height: 2, Grid: "##,##", Status: model.MotifValid,
	}); err != nil {
		t.Fatalf("motif: %v", err)
	}
	n, err := ValidMotifsOfSample(st.Motifs, smp.ID)
	if err != nil || n != 1 {
		t.Fatalf("valid motifs = %d err=%v", n, err)
	}
}
