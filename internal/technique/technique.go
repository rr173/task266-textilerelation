// Package technique 负责经纬交错规则与染料证据的工艺验证。
//
// 纺织史研究的关键问题：两个纹样"看起来像"不等于"同源"。
// 本包从织造技法链入手验证：
//   - 交错规则（Interlacing）：结构族 + 浮长，规则相反/结构族不同 → 冲突；
//   - 染料证据（Dye）：天然/合成类别、色素、碳同位素比值；
//   - 织造核验（VerifyWeave）：样本宣称的织法类别与经纬参数是否自洽。
//
// 工艺冲突是把"拓扑相似的候选"降级为"视觉巧合"的核心判据，
// 对应端到端场景：两纹样外观相似但经纬交错规则相反 → 系统否决传承候选。
package technique

import (
	"fmt"

	"task266-textilerelation/internal/model"
)

// Assessment 是一次工艺验证的综合结论。
type Assessment struct {
	WeaveCompat string `json:"weave_compat"` // compatible/partial/conflict/unknown
	DyeCompat   string `json:"dye_compat"`   // compatible/unknown/conflict
	Summary     string `json:"summary"`
}

// Assess 结合两个样本的工艺特征给出织造与染料双维兼容性。
//
// 技法数据缺失（任一样本无工艺特征）时返回 ErrTechnique，
// 由调用方决定按 unknown 处理或拒绝比较。
func Assess(a, b *model.TechniqueFeature) (*Assessment, error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("%w: technique feature required", model.ErrTechnique)
	}
	weave := model.CompatRule(a.WeaveClass, a.Interlacing, b.WeaveClass, b.Interlacing)
	dye := model.DyeCompatRule(a.DyeClass, a.DyePigment, a.CarbonRatio,
		b.DyeClass, b.DyePigment, b.CarbonRatio)

	summary := buildSummary(a, b, weave, dye)
	return &Assessment{WeaveCompat: weave, DyeCompat: dye, Summary: summary}, nil
}

// buildSummary 生成工艺验证的人类可读摘要。
func buildSummary(a, b *model.TechniqueFeature, weave, dye string) string {
	return fmt.Sprintf(
		"织造 %s(%s) 对 %s(%s)：交错%s；染料 %s(%s) 对 %s(%s)：%s",
		a.WeaveClass, a.Interlacing, b.WeaveClass, b.Interlacing, weave,
		a.DyeClass, a.DyePigment, b.DyeClass, b.DyePigment, dye)
}

// VerifyWeave 核验样本工艺特征自洽性：织法类别与经纬密度、浮长相匹配。
//
//   - plain：经密/纬密应接近（比值 ≤ 1.5），否则可能误判织法；
//   - twill：要求斜纹方向合法（left/right）且浮长以 2/1、3/1 常见形态；
//   - satin：浮长以 5/2、8/3 等长浮长形态；
//   - compound：允许经纬密度差异大（纬起花）。
func VerifyWeave(f *model.TechniqueFeature, warpDensity, weftDensity int) (bool, string) {
	ratio := float64(warpDensity) / float64(weftDensity)
	if ratio < 1 {
		ratio = 1 / ratio
	}
	switch f.WeaveClass {
	case model.WeavePlain:
		if ratio > 1.5 {
			return false, fmt.Sprintf("平纹要求经纬密度接近，实际比 %.2f", ratio)
		}
		return true, "平纹经纬密度自洽"
	case model.WeaveTwill:
		if f.TwillDirection != "left" && f.TwillDirection != "right" {
			return false, fmt.Sprintf("斜纹方向非法: %s", f.TwillDirection)
		}
		if f.Interlacing != model.RuleTwill21 && f.Interlacing != model.RuleTwill31 {
			return false, fmt.Sprintf("斜纹浮长异常: %s", f.Interlacing)
		}
		return true, "斜纹方向与浮长自洽"
	case model.WeaveSatin:
		if f.Interlacing != model.RuleSatin52 && f.Interlacing != model.RuleSatin83 {
			return false, fmt.Sprintf("缎纹浮长异常: %s", f.Interlacing)
		}
		return true, "缎纹长浮长自洽"
	case model.WeaveCompound:
		return true, "重组织允许经纬密度大差异"
	default:
		return false, fmt.Sprintf("未知织法: %s", f.WeaveClass)
	}
}
