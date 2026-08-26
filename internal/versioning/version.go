// Package versioning 管理工艺传承关系图的发布版本。
//
// 版本状态机：draft → shared → frozen → superseded。
//   - draft：可增删关系、可改摘要；
//   - shared：对外共享，关系归属锁定，仅可冻结；
//   - frozen：不可变快照，保留出处说明与反证摘要，只能被新版本替代；
//   - superseded：已被替代，仅保留只读历史。
//
// 冻结后任何修改（加关系/改摘要）返回 ErrFrozen。
// 发布前必须通过 cycle.Detect 出处循环检测：关系图成环（A→B→A）时
// 拒绝发布，防止传承证据出现自指矛盾。
package versioning

import (
	"fmt"

	"task266-textilerelation/internal/model"
)

// Transition 校验版本状态流转。
func Transition(from, to string) error {
	switch from {
	case model.VersionDraft:
		if to != model.VersionShared {
			return fmt.Errorf("%w: draft can only -> shared", model.ErrState)
		}
	case model.VersionShared:
		if to != model.VersionFrozen {
			return fmt.Errorf("%w: shared can only -> frozen", model.ErrState)
		}
	case model.VersionFrozen:
		if to != model.VersionSuperseded {
			return fmt.Errorf("%w: frozen can only -> superseded", model.ErrState)
		}
	case model.VersionSuperseded:
		return fmt.Errorf("%w: superseded is terminal", model.ErrState)
	default:
		return fmt.Errorf("%w: unknown version status %q", model.ErrState, from)
	}
	return nil
}

// CanMutate 判断版本当前是否允许增删关系（仅草稿可变更）。
func CanMutate(v *model.RelationVersion) error {
	if v.Status != model.VersionDraft {
		return fmt.Errorf("%w: version %d status=%s", model.ErrFrozen, v.ID, v.Status)
	}
	return nil
}

// BuildProvenanceNote 由关系列表与反证汇总生成版本出处说明。
func BuildProvenanceNote(relationCount int, counterEvidenceNote string) string {
	return fmt.Sprintf("收录 %d 条工艺关系；%s", relationCount, counterEvidenceNote)
}
