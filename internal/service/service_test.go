package service

import (
	"path/filepath"
	"testing"

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
