package technique

import (
	"testing"

	"task266-textilerelation/internal/model"
)

func TestCompatRule(t *testing.T) {
	// 平纹 vs 斜纹 → 结构族不同 → conflict
	if got := model.CompatRule(model.WeavePlain, model.RulePlain11, model.WeaveTwill, model.RuleTwill31); got != "conflict" {
		t.Errorf("plain vs twill = %s, want conflict", got)
	}
	// 斜纹同浮长 → compatible
	if got := model.CompatRule(model.WeaveTwill, model.RuleTwill31, model.WeaveTwill, model.RuleTwill31); got != "compatible" {
		t.Errorf("twill same = %s, want compatible", got)
	}
	// 同族不同浮长 → partial
	if got := model.CompatRule(model.WeaveTwill, model.RuleTwill21, model.WeaveTwill, model.RuleTwill31); got != "partial" {
		t.Errorf("twill diff = %s, want partial", got)
	}
	// 缎纹与斜纹同族（floats）不同浮长 → partial
	if got := model.CompatRule(model.WeaveSatin, model.RuleSatin52, model.WeaveTwill, model.RuleTwill31); got != "partial" {
		t.Errorf("satin vs twill = %s, want partial", got)
	}
}

func TestDyeCompatRule(t *testing.T) {
	// 天然 vs 合成 → conflict
	if got := model.DyeCompatRule("natural", "茜草", -18.5, "synthetic", "茜草", -28.0); got != "conflict" {
		t.Errorf("natural vs synthetic = %s, want conflict", got)
	}
	// 同天然，碳比差 0.2 → compatible
	if got := model.DyeCompatRule("natural", "靛蓝", -25.0, "natural", "靛蓝", -24.8); got != "compatible" {
		t.Errorf("same natural = %s, want compatible", got)
	}
	// 同天然，碳比差 2.0 → conflict
	if got := model.DyeCompatRule("natural", "茜草", -18.5, "natural", "茜草", -20.5); got != "conflict" {
		t.Errorf("carbon gap = %s, want conflict", got)
	}
}

func TestVerifyWeave(t *testing.T) {
	f := &model.TechniqueFeature{
		WeaveClass: model.WeaveTwill, Interlacing: model.RuleTwill31, TwillDirection: "right",
	}
	ok, _ := VerifyWeave(f, 24, 22)
	if !ok {
		t.Error("valid twill should pass")
	}
	bad := &model.TechniqueFeature{
		WeaveClass: model.WeaveTwill, Interlacing: model.RuleTwill31, TwillDirection: "diagonal",
	}
	if ok, _ := VerifyWeave(bad, 24, 22); ok {
		t.Error("invalid twill direction should fail")
	}
}

func TestAssessMissingTechnique(t *testing.T) {
	if _, err := Assess(nil, nil); err == nil {
		t.Fatal("expected error for nil features")
	}
}
