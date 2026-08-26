package technique

import (
	"fmt"
	"strconv"

	"task266-textilerelation/internal/model"
)

// dyePigmentFamily 色素归族：同一色素族内的天然染料共享染源植物/矿物。
var dyePigmentFamily = map[string]string{
	"茜草":  "madder",
	"红土":  "madder",
	"靛蓝":  "indigo",
	"菘蓝":  "indigo",
	"黄栌":  "weld",
	"姜黄":  "weld",
	"苏木":  "brazilwood",
	"五倍子": "gall",
	"黄檗":  "berberine",
}

// PigmentFamily 返回色素的染源族；未知色素返回空串。
func PigmentFamily(pigment string) string {
	return dyePigmentFamily[pigment]
}

// DyeEvidence 汇总一个样本的染料证据强度。
type DyeEvidence struct {
	Class       string  `json:"class"`
	Pigment     string  `json:"pigment"`
	Colorfast   int     `json:"colorfastness"`
	CarbonRatio float64 `json:"carbon_ratio"`
	// Strength 证据强度 0~1：色牢度高、色素可归族、碳比已测则更强。
	Strength float64 `json:"strength"`
}

// AssessDye 单独评估样本染料证据强度（用于版本出处说明）。
func AssessDye(f *model.TechniqueFeature) DyeEvidence {
	s := 0.3
	if f.Colorfastness >= 4 {
		s += 0.3
	}
	if PigmentFamily(f.DyePigment) != "" {
		s += 0.2
	}
	if f.CarbonRatio != 0 {
		s += 0.2
	}
	if s > 1 {
		s = 1
	}
	return DyeEvidence{
		Class:       f.DyeClass,
		Pigment:     f.DyePigment,
		Colorfast:   f.Colorfastness,
		CarbonRatio: f.CarbonRatio,
		Strength:    s,
	}
}

// DyeSummary 输出染料证据的可读说明。
func DyeSummary(f *model.TechniqueFeature) string {
	ev := AssessDye(f)
	fam := PigmentFamily(f.DyePigment)
	base := ev.Class + "染料"
	if fam != "" {
		base += "（染源族 " + fam + "）"
	}
	return base + " 色素" + f.DyePigment + " 色牢度" +
		strconv.Itoa(f.Colorfastness) + "级 证据强度" +
		fmt.Sprintf("%.2f", ev.Strength)
}
