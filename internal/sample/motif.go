package sample

import (
	"fmt"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/store"
	"task266-textilerelation/internal/topology"
)

// AddMotif 向样本添加纹样单元（pending 态，需解析后转为 valid）。
func AddMotif(samples *store.SampleStore, motifs *store.MotifStore, sampleID int64,
	name string, originX, originY, width, height int, grid string) (*model.MotifUnit, error) {
	sample, err := samples.Get(sampleID)
	if err != nil {
		return nil, err
	}
	if sample.Status == model.BatchSealed {
		return nil, fmt.Errorf("%w: sample sealed, cannot add motif", model.ErrFrozen)
	}
	if err := model.ValidateMotif(sample, name, originX, originY, width, height, grid); err != nil {
		return nil, err
	}
	return motifs.Create(&model.MotifUnit{
		SampleID: sampleID,
		Name:     name,
		OriginX:  originX,
		OriginY:  originY,
		Width:    width,
		Height:   height,
		Grid:     grid,
		Status:   model.MotifPending,
	})
}

// ParseMotif 解析纹样单元：抽取网格 → 检测周期 → 检测对称 → 标记 valid。
//
// 解析失败（网格格式错误）时返回 ErrInvalid；解析成功后单元状态变为 valid，
// 周期与对称回填，作为拓扑比较与传承判定的结构指纹。
func ParseMotif(motifs *store.MotifStore, motifID int64) (*model.MotifUnit, error) {
	u, err := motifs.Get(motifID)
	if err != nil {
		return nil, err
	}
	if u.Status != model.MotifPending {
		return nil, fmt.Errorf("%w: motif %d already parsed (status=%s)", model.ErrState, u.ID, u.Status)
	}
	g, err := topology.Parse(u.Grid)
	if err != nil {
		return nil, err
	}
	periodX, periodY := topology.DetectPeriod(g)
	sym := topology.DetectSymmetry(g)
	if err := motifs.UpdateParsed(u.ID, periodX, periodY, sym, model.MotifValid); err != nil {
		return nil, err
	}
	return motifs.Get(u.ID)
}

// ExcludeMotif 将单元排除出传承证据链。
//
// reason 决定语义：patch 表示后期补片（保留记录但不作证据），
// excluded 表示彻底排除。两者都会阻止该单元参与候选生成。
func ExcludeMotif(motifs *store.MotifStore, motifID int64, asPatch bool) (*model.MotifUnit, error) {
	u, err := motifs.Get(motifID)
	if err != nil {
		return nil, err
	}
	if u.Status == model.MotifExcluded {
		return nil, fmt.Errorf("%w: motif %d already excluded", model.ErrState, u.ID)
	}
	status := model.MotifExcluded
	if asPatch {
		status = model.MotifPatch
	}
	if err := motifs.UpdateStatus(u.ID, status); err != nil {
		return nil, err
	}
	return motifs.Get(u.ID)
}

// ValidMotifsOfSample 统计样本内有效单元数量（批次状态机约束用）。
func ValidMotifsOfSample(motifs *store.MotifStore, sampleID int64) (int, error) {
	list, err := motifs.ListBySample(sampleID)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, u := range list {
		if u.Status == model.MotifValid {
			n++
		}
	}
	return n, nil
}

// ValidMotifs 返回全部有效单元（拓扑比较的候选源）。
func ValidMotifs(motifs *store.MotifStore) ([]*model.MotifUnit, error) {
	list, err := motifs.List()
	if err != nil {
		return nil, err
	}
	out := make([]*model.MotifUnit, 0, len(list))
	for _, u := range list {
		if u.Status == model.MotifValid {
			out = append(out, u)
		}
	}
	return out, nil
}
