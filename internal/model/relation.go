package model

import "time"

// RelationCandidate 是两个纹样单元之间的工艺传承关系候选。
//
// 判定由拓扑比较（TopoSimilarity）与工艺验证（WeaveCompat/DyeCompat）共同得出；
// 状态机：candidate → partial → conflict/confirmed/rejected。
type RelationCandidate struct {
	ID             int64     `json:"id"`
	FromUnitID     int64     `json:"from_unit_id"`
	ToUnitID       int64     `json:"to_unit_id"`
	TopoSimilarity float64   `json:"topo_similarity"` // 0~1
	WeaveCompat    string    `json:"weave_compat"`    // compatible/partial/conflict
	DyeCompat      string    `json:"dye_compat"`      // compatible/unknown/conflict
	Verdict        string    `json:"verdict"`         // candidate/partial/conflict/confirmed/rejected
	Summary        string    `json:"summary"`
	VersionID      int64     `json:"version_id"` // 0=未收录进版本
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CounterEvidence 是针对某个关系候选的反证：补片、出处反证或层位证据。
type CounterEvidence struct {
	ID          int64     `json:"id"`
	RelationID  int64     `json:"relation_id"`
	Kind        string    `json:"kind"` // patch/opposing_source/stratigraphy
	Description string    `json:"description"`
	Ref         string    `json:"ref"` // 文献/档案/报告编号
	CreatedAt   time.Time `json:"created_at"`
}

// RelationVersion 是工艺传承关系图的发布版本。
//
// 状态机：draft → shared → frozen → superseded；
// frozen 之后关系归属不可再变更，出处与反证快照被保留。
type RelationVersion struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	Summary        string    `json:"summary"`
	RelationIDs    string    `json:"relation_ids"`    // 逗号分隔
	ProvenanceNote string    `json:"provenance_note"` // 出处说明与反证摘要
	CreatedAt      time.Time `json:"created_at"`
	FrozenAt       time.Time `json:"frozen_at"`
	SupersededBy   int64     `json:"superseded_by"` // 0=未被替代
}

// BatchStatus 批次状态集合。
const (
	BatchPending   = "pending"   // 整理中
	BatchReady     = "ready"     // 待比较
	BatchReviewing = "reviewing" // 待复核
	BatchPublished = "published" // 已发布
	BatchSealed    = "sealed"    // 封存
)

// MotifStatus 纹样单元状态集合。
const (
	MotifPending  = "pending"  // 待解析
	MotifValid    = "valid"    // 有效
	MotifPatch    = "patch"    // 后期补片
	MotifExcluded = "excluded" // 排除
)

// RelationVerdict 关系候选裁决集合。
const (
	VerdictCandidate = "candidate" // 候选
	VerdictPartial   = "partial"   // 部分支持
	VerdictConflict  = "conflict"  // 冲突（相似但不同源）
	VerdictConfirmed = "confirmed" // 确认
	VerdictRejected  = "rejected"  // 否决
)

// VersionStatus 关系版本状态集合。
const (
	VersionDraft      = "draft"      // 草稿
	VersionShared     = "shared"     // 共享
	VersionFrozen     = "frozen"     // 冻结
	VersionSuperseded = "superseded" // 替代
)
