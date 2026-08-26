package relation

import (
	"fmt"

	"task266-textilerelation/internal/model"
)

// VerdictRule 定义候选状态机的合法流转。
//
//	candidate → confirmed / partial / conflict / rejected
//	partial   → confirmed / conflict / rejected
//	conflict  → rejected / partial（证据更正）
//	confirmed → rejected（强反证推翻）
//	rejected  为终态。
var VerdictRule = map[string][]string{
	model.VerdictCandidate: {model.VerdictConfirmed, model.VerdictPartial, model.VerdictConflict, model.VerdictRejected},
	model.VerdictPartial:   {model.VerdictConfirmed, model.VerdictConflict, model.VerdictRejected},
	model.VerdictConflict:  {model.VerdictRejected, model.VerdictPartial},
	model.VerdictConfirmed: {model.VerdictRejected},
	model.VerdictRejected:  {},
}

// CanTransition 判断裁决流转是否合法。
func CanTransition(from, to string) bool {
	for _, next := range VerdictRule[from] {
		if next == to {
			return true
		}
	}
	return false
}

// ApplyVerdict 校验并应用人工裁决；非法流转返回 ErrState。
func ApplyVerdict(rel *model.RelationCandidate, verdict, summary string) error {
	if !CanTransition(rel.Verdict, verdict) {
		return fmt.Errorf("%w: cannot %s -> %s", model.ErrState, rel.Verdict, verdict)
	}
	if summary == "" {
		summary = defaultSummary(verdict)
	}
	return nil
}

// defaultSummary 为裁决生成默认说明。
func defaultSummary(verdict string) string {
	switch verdict {
	case model.VerdictConfirmed:
		return "人工复核确认：拓扑、织造与染料证据一致，工艺传承关系成立"
	case model.VerdictRejected:
		return "人工复核否决：证据不足以支持工艺传承"
	case model.VerdictConflict:
		return "工艺冲突成立：交错规则相反或结构族不同，判为视觉巧合"
	case model.VerdictPartial:
		return "部分支持：核心证据一致但存在待澄清项"
	default:
		return "裁决更新"
	}
}
