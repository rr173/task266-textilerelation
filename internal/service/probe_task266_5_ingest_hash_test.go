package service

import (
	"testing"

	"task266-textilerelation/internal/sample"
)

func TestIngestHashIncludesMotifGrids(t *testing.T) {
	svc := newTestService(t)
	batch, err := svc.CreateBatch("hash-batch", "")
	if err != nil {
		t.Fatal(err)
	}
	base := sample.IngestInput{
		BatchID: batch.ID, Name: "frag-x", Provenance: "site-x",
		WarpCount: 8, WeftCount: 8, WarpDensity: 16, WeftDensity: 14,
		MotifGrids: map[string]string{"diamond": "##.#,.##."},
	}
	first, created, err := svc.ImportSample(base)
	if err != nil || !created {
		t.Fatalf("first ingest: created=%v err=%v", created, err)
	}
	other := base
	other.MotifGrids = map[string]string{"diamond": "##,##"}
	second, created2, err := svc.ImportSample(other)
	if err != nil {
		t.Fatal(err)
	}
	if !created2 {
		t.Fatal("different motif grids must create a new sample")
	}
	if second.ID == first.ID {
		t.Fatalf("different grids must not hit same sample id %d", first.ID)
	}
}
