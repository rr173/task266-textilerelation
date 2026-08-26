package topology

import (
	"math"
	"testing"

	"task266-textilerelation/internal/model"
)

func TestParseAndDetectPeriod(t *testing.T) {
	g, err := Parse("##.#,.##.,##.#,.##.")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if g.Rows != 4 || g.Cols != 4 {
		t.Fatalf("size = %dx%d", g.Rows, g.Cols)
	}
	px, py := DetectPeriod(g)
	// 该网格第 0/2 行相同、第 1/3 行相同 → 垂直周期 2；水平无周期。
	if py != 2 {
		t.Errorf("periodY = %d, want 2", py)
	}
	if px != 0 {
		t.Errorf("periodX = %d, want 0", px)
	}
}

func TestDetectSymmetry(t *testing.T) {
	cases := []struct {
		grid string
		want string
	}{
		{"##,##", "hv"},   // 2x2 全对称
		{"##,..", "h"},    // 每行内部水平对称
		{"#.,.#", "rot180"},
		{"##,#.", "none"}, // 非对称
	}
	for _, c := range cases {
		g, err := Parse(c.grid)
		if err != nil {
			t.Fatalf("parse %s: %v", c.grid, err)
		}
		if got := DetectSymmetry(g); got != c.want {
			t.Errorf("DetectSymmetry(%s) = %s, want %s", c.grid, got, c.want)
		}
	}
}

func TestCompareMotifs(t *testing.T) {
	a := &model.MotifUnit{ID: 1, Grid: "##.#,.##.,##.#,.##.", Status: model.MotifValid, PeriodX: 0, PeriodY: 2, Symmetry: "none"}
	b := &model.MotifUnit{ID: 2, Grid: "##.#,.###,##.#,.###", Status: model.MotifValid, PeriodX: 0, PeriodY: 2, Symmetry: "none"}
	comp, err := CompareMotifs(a, b)
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if comp.OverlapRate != 14.0/16.0 {
		t.Errorf("overlap = %v, want %v", comp.OverlapRate, 14.0/16.0)
	}
	if math.Abs(comp.TopoSimilarity-0.9125) > 1e-9 {
		t.Errorf("topo similarity = %v, want 0.9125", comp.TopoSimilarity)
	}
}

func TestCompareMotifsRejectsInvalid(t *testing.T) {
	a := &model.MotifUnit{ID: 1, Grid: "##.#", Status: model.MotifPending}
	b := &model.MotifUnit{ID: 2, Grid: "##.#", Status: model.MotifValid}
	if _, err := CompareMotifs(a, b); err == nil {
		t.Fatal("expected error for pending motif")
	}
}

func TestSymmetryCompatible(t *testing.T) {
	if !SymmetryCompatible("h", "rot180") {
		t.Error("h and rot180 should be same family")
	}
	if SymmetryCompatible("none", "h") {
		t.Error("none and h should differ")
	}
	if !SymmetryCompatible("none", "none") {
		t.Error("none-none should match")
	}
}
