package versioning

import (
	"errors"
	"strings"
	"testing"

	"task266-textilerelation/internal/model"
)

func TestDetectCycle(t *testing.T) {
	// 无环：A→B→C
	if err := DetectCycle([][2]int64{{1, 2}, {2, 3}}); err != nil {
		t.Fatalf("unexpected cycle: %v", err)
	}
	// 自环
	if err := DetectCycle([][2]int64{{1, 1}}); err == nil {
		t.Fatal("expected self cycle")
	}
	// 双环
	err := DetectCycle([][2]int64{{1, 2}, {2, 1}})
	if err == nil {
		t.Fatal("expected 2-cycle")
	}
	if !strings.Contains(err.Error(), "1 -> 2") {
		t.Fatalf("cycle path missing: %v", err)
	}
	// 三环
	if err := DetectCycle([][2]int64{{1, 2}, {2, 3}, {3, 1}}); err == nil {
		t.Fatal("expected 3-cycle")
	}

	// 两个单元互相确认传承（A→B 且 B→A）必须成环拒绝
	err = DetectCycle([][2]int64{{1, 2}, {2, 1}})
	if err == nil {
		t.Fatal("expected mutual-confirmation cycle")
	}
	if !errors.Is(err, model.ErrCycle) {
		t.Fatalf("want ErrCycle, got %v", err)
	}
}

func TestTransition(t *testing.T) {
	if err := Transition(model.VersionDraft, model.VersionShared); err != nil {
		t.Errorf("draft->shared: %v", err)
	}
	if err := Transition(model.VersionShared, model.VersionFrozen); err != nil {
		t.Errorf("shared->frozen: %v", err)
	}
	if err := Transition(model.VersionFrozen, model.VersionSuperseded); err != nil {
		t.Errorf("frozen->superseded: %v", err)
	}
	if err := Transition(model.VersionDraft, model.VersionFrozen); err == nil {
		t.Error("draft->frozen should fail")
	}
	if err := Transition(model.VersionSuperseded, model.VersionDraft); err == nil {
		t.Error("superseded should be terminal")
	}
}
