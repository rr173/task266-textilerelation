// Package sample 负责样本批次的导入组织与状态机。
//
// 批次状态机：pending（整理中）→ ready（待比较）→ reviewing（待复核）
// → published（已发布）→ sealed（封存）。
// 流转约束：
//   - pending → ready 要求批次内至少 1 个样本；
//   - ready → reviewing 要求批次内所有样本均已解析出有效纹样单元；
//   - published → sealed 为终态，封存后不再接受样本录入。
//
// 样本导入按内容哈希幂等：相同结构（名称/出处/经纬参数）重复导入
// 返回既有样本，不会产生重复记录（对应"样本哈希幂等"持久化要求）。
package sample

import (
	"fmt"

	"task266-textilerelation/internal/model"
)

// BatchTransitions 批次状态机的合法流转表。
var BatchTransitions = map[string][]string{
	model.BatchPending:   {model.BatchReady},
	model.BatchReady:     {model.BatchReviewing},
	model.BatchReviewing: {model.BatchPublished},
	model.BatchPublished: {model.BatchSealed},
	model.BatchSealed:    {},
}

// NextBatchStatus 依据当前状态与约束计算下一个可流转状态。
func NextBatchStatus(current string, sampleCount, validMotifCount int) (string, error) {
	nexts := BatchTransitions[current]
	if len(nexts) == 0 {
		return "", fmt.Errorf("%w: batch %s is terminal", model.ErrState, current)
	}
	next := nexts[0]
	switch next {
	case model.BatchReady:
		if sampleCount == 0 {
			return "", fmt.Errorf("%w: need at least 1 sample before ready", model.ErrState)
		}
	case model.BatchReviewing:
		if validMotifCount == 0 {
			return "", fmt.Errorf("%w: need parsed valid motifs before reviewing", model.ErrState)
		}
	}
	return next, nil
}
