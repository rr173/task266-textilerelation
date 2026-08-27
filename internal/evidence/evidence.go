// Package evidence 管理工艺传承关系的反证，并评估反证对裁决的影响。
//
// 反证类型：
//   - patch（补片）：某样本的纹样是后期缝补，不能作为传承证据；
//   - opposing_source（出处反证）：两样本出处年代/遗址冲突；
//   - stratigraphy（层位）：出土层位矛盾。
//
// 反证数量与类型决定证据强度：任一反证存在即要求候选降级复核；
// 出处反证（opposing_source）直接否决传承关系。
package evidence

import (
	"fmt"
	"sort"
	"strings"

	"task266-textilerelation/internal/model"
)

// Impact 是反证对候选的影响评估。
type Impact struct {
	Count       int      `json:"count"`
	Kinds       []string `json:"kinds"`
	Downgrade   bool     `json:"downgrade"`   // 需要降级复核
	Reject      bool     `json:"reject"`      // 直接否决
	Suggested   string   `json:"suggested"`   // 建议裁决
	Explanation string   `json:"explanation"` // 说明
}

// AssessImpact 评估反证集合对候选的影响。
func AssessImpact(items []*model.CounterEvidence, currentVerdict string) Impact {
	imp := Impact{Count: len(items)}
	kinds := map[string]bool{}
	for _, it := range items {
		kinds[it.Kind] = true
		switch it.Kind {
		case "patch":
			imp.Downgrade = true
		case "opposing_source", "stratigraphy":
			imp.Reject = true
		}
	}
	for k := range kinds {
		imp.Kinds = append(imp.Kinds, k)
	}
	sort.Strings(imp.Kinds)

	switch {
	case imp.Reject:
		imp.Suggested = model.VerdictRejected
		imp.Explanation = "存在出处/层位反证，工艺传承关系不成立"
	case imp.Downgrade:
		imp.Suggested = model.VerdictPartial
		imp.Explanation = "存在补片反证：纹样可能为后期缝补，传承证据降级"
	case currentVerdict == model.VerdictConfirmed && len(items) == 0:
		imp.Suggested = model.VerdictConfirmed
		imp.Explanation = "无反证，确认裁决维持"
	default:
		imp.Suggested = currentVerdict
		imp.Explanation = "反证不足，维持当前裁决"
	}
	return imp
}

// ProvenanceNote 汇总一批反证为版本出处说明文本。
func ProvenanceNote(items []*model.CounterEvidence) string {
	if len(items) == 0 {
		return "无反证记录"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "反证 %d 条：", len(items))
	for i, it := range items {
		if i > 0 {
			sb.WriteString("；")
		}
		fmt.Fprintf(&sb, "[%s] %s（%s）", it.Kind, it.Description, it.Ref)
	}
	return sb.String()
}

// ValidateKind 校验反证类型合法性。
func ValidateKind(kind string) error {
	switch kind {
	case "patch", "opposing_source", "stratigraphy":
		return nil
	default:
		return fmt.Errorf("%w: unknown evidence kind %q", model.ErrInvalid, kind)
	}
}
