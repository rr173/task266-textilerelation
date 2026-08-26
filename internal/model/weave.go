package model

import "fmt"

// 织造组织学（Weave structure）常量。
//
// 纺织史研究中，外观相似的纹样可能来自完全不同的织造技法链：
// 平纹（1/1 经纬一上一下）、斜纹（2/1、3/1 浮长斜向连续）、
// 缎纹（5/2、8/3 长浮长稀疏交织）、重组织（compound 双层/纬起花）。
// 交错规则相反或结构族不同，是"相似但不同源"（视觉巧合）的核心判据。

// WeaveClass 织法类别集合。
const (
	WeavePlain    = "plain"    // 平纹 1/1
	WeaveTwill    = "twill"    // 斜纹
	WeaveSatin    = "satin"    // 缎纹
	WeaveCompound = "compound" // 重组织/纬起花
)

// 常见浮长规则（经浮长/纬浮长）。
const (
	RulePlain11   = "1/1"
	RuleTwill21   = "2/1"
	RuleTwill31   = "3/1"
	RuleSatin52   = "5/2"
	RuleSatin83   = "8/3"
	RuleCompound  = "compound"
)

// WeaveFamily 返回织法所属结构族：twill 与 satin 归入斜向浮长族，plain 单独成族。
func WeaveFamily(weaveClass string) string {
	switch weaveClass {
	case WeavePlain:
		return "plain"
	case WeaveTwill, WeaveSatin:
		return "floats"
	case WeaveCompound:
		return "compound"
	default:
		return "unknown"
	}
}

// CompatRule 比较两个交错规则：结构族 + 浮长一致为 compatible；
// 结构族一致但浮长不同为 partial；结构族不同为 conflict。
func CompatRule(aClass, aRule, bClass, bRule string) string {
	famA, famB := WeaveFamily(aClass), WeaveFamily(bClass)
	if famA == "unknown" || famB == "unknown" {
		return "unknown"
	}
	if famA != famB {
		return "conflict"
	}
	if aRule == bRule {
		return "compatible"
	}
	return "partial"
}

// DescribeRule 将交错规则转为可读描述（用于候选摘要与版本出处说明）。
func DescribeRule(weaveClass, interlacing, direction string) string {
	switch weaveClass {
	case WeavePlain:
		return "平纹 1/1 一上一下"
	case WeaveTwill:
		d := "左斜"
		if direction == "right" {
			d = "右斜"
		}
		return fmt.Sprintf("斜纹 %s %s浮长连续", interlacing, d)
	case WeaveSatin:
		return fmt.Sprintf("缎纹 %s 长浮长稀疏交织", interlacing)
	case WeaveCompound:
		return "重组织 纬起花 多层"
	default:
		return interlacing
	}
}

// DyeCompatRule 依据染料证据给出兼容性：天然/合成类别冲突即为 conflict，
// 同类别下碳同位素差异超过阈值视为 conflict，色素一致视为 compatible。
func DyeCompatRule(aClass, aPigment string, aRatio float64, bClass, bPigment string, bRatio float64) string {
	if aClass != bClass {
		return "conflict"
	}
	if aClass == "synthetic" {
		// 合成染料碳源均一，同位素无鉴别力，但色素必须一致。
		if aPigment == bPigment {
			return "compatible"
		}
		return "unknown"
	}
	diff := aRatio - bRatio
	if diff < 0 {
		diff = -diff
	}
	if diff > 1.5 {
		return "conflict"
	}
	if aPigment == bPigment {
		return "compatible"
	}
	return "unknown"
}
