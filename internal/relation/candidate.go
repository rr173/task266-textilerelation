// Package relation 负责工艺传承关系候选的生成与裁决。
//
// 候选生成流程：
//  1. 拓扑比较（topology.CompareMotifs）给出综合相似度；
//  2. 工艺验证（technique.Assess）给出织造/染料兼容性；
//  3. 裁决规则：
//     - 拓扑相似度 < 0.5 → 不生成候选（返回 rejected 语义）；
//     - 拓扑相似且织造兼容（compatible）→ confirmed（强传承证据）；
//     - 拓扑相似但织造冲突（conflict，交错规则相反/结构族不同）→ conflict
//       ——"相似但不同源"视觉巧合，端到端场景核心；
//     - 其余 → candidate / partial，等待人工复核。
//
// 裁决支持版本校验（UpdateVerdict 条件更新），同一候选并发复核时
// 后写者会收到 StateMismatchError，保证裁决链不被打乱。
package relation

import (
	"fmt"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/technique"
	"task266-textilerelation/internal/topology"
)

// GenerateInput 是候选生成的入参：两个已解析单元及其样本的工艺特征。
type GenerateInput struct {
	From *model.MotifUnit
	To   *model.MotifUnit
	// FromTech/ToTech 允许为 nil；nil 时织造维度按 unknown 处理（不阻断拓扑候选）。
	FromTech *model.TechniqueFeature
	ToTech   *model.TechniqueFeature
}

// Generated 是候选生成结果。
type Generated struct {
	TopoSimilarity float64
	WeaveCompat    string
	DyeCompat      string
	Verdict        string
	Summary        string
}

// Generate 依据拓扑与工艺证据生成候选裁决。
func Generate(in GenerateInput) (*Generated, error) {
	if in.From == nil || in.To == nil {
		return nil, fmt.Errorf("%w: motifs required", model.ErrInvalid)
	}
	comp, err := topology.CompareMotifs(in.From, in.To)
	if err != nil {
		return nil, err
	}

	weaveCompat := "unknown"
	dyeCompat := "unknown"
	if in.FromTech != nil && in.ToTech != nil {
		as, aerr := technique.Assess(in.FromTech, in.ToTech)
		if aerr == nil {
			weaveCompat = as.WeaveCompat
			dyeCompat = as.DyeCompat
		}
	}

	verdict := decideVerdict(comp.TopoSimilarity, weaveCompat, dyeCompat)
	summary := fmt.Sprintf(
		"%s；%s",
		comp.Reason,
		techniqueSummary(weaveCompat, dyeCompat),
	)
	return &Generated{
		TopoSimilarity: comp.TopoSimilarity,
		WeaveCompat:    weaveCompat,
		DyeCompat:      dyeCompat,
		Verdict:        verdict,
		Summary:        summary,
	}, nil
}

// decideVerdict 把拓扑相似度与工艺兼容性映射为裁决。
func decideVerdict(sim float64, weave, dye string) string {
	switch {
	case sim < 0.5:
		return model.VerdictRejected
	case sim >= 0.7 && weave == "compatible":
		return model.VerdictConfirmed
	case sim >= 0.7 && weave == "conflict":
		// 外观高度相似但经纬交错规则相反/结构族不同 → 视觉巧合。
		return model.VerdictConflict
	case sim >= 0.7 && weave == "partial":
		return model.VerdictPartial
	case sim >= 0.5:
		return model.VerdictCandidate
	default:
		return model.VerdictRejected
	}
}

func techniqueSummary(weave, dye string) string {
	return fmt.Sprintf("织造兼容=%s，染料兼容=%s", weave, dye)
}

// ResolveConflict 端到端场景专用：人工复核"外观相似但工艺冲突"的候选。
//
// 复核结论确认工艺冲突成立时，候选被否决（rejected）并保留冲突摘要；
// 若研究者补充证据推翻冲突（如交错规则读数更正），可降级为部分支持。
func ResolveConflict(rel *model.RelationCandidate, oppose bool) (string, string) {
	if oppose {
		return model.VerdictRejected,
			"复核确认：经纬交错规则相反，判为视觉巧合，否决工艺传承"
	}
	return model.VerdictPartial,
		"复核更正：交错规则读数更新，工艺冲突解除，转为部分支持"
}
