// Package service 编排样本导入、拓扑比较、工艺验证、反证管理与版本发布，
// 是 HTTP 层与业务包之间的唯一编排入口。
package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"task266-textilerelation/internal/evidence"
	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/relation"
	"task266-textilerelation/internal/sample"
	"task266-textilerelation/internal/store"
	"task266-textilerelation/internal/technique"
	"task266-textilerelation/internal/versioning"
)

// Service 聚合持久化仓库与领域服务。
type Service struct {
	Store  *store.Store
	Ingest *sample.IngestService
}

// New 构造服务。
func New(st *store.Store) *Service {
	return &Service{
		Store:  st,
		Ingest: sample.NewIngestService(st),
	}
}

// ---- 批次 ----

// CreateBatch 新建批次。
func (s *Service) CreateBatch(name, note string) (*model.Batch, error) {
	return s.Store.Batches.Create(name, note)
}

// ListBatches 列出批次。
func (s *Service) ListBatches() ([]*model.Batch, error) {
	return s.Store.Batches.List()
}

// GetBatch 查询批次详情（含样本数）。
func (s *Service) GetBatch(id int64) (*model.Batch, error) {
	return s.Store.Batches.Get(id)
}

// AdvanceBatch 推进批次状态机。
func (s *Service) AdvanceBatch(id int64) (*model.Batch, error) {
	b, err := s.Store.Batches.Get(id)
	if err != nil {
		return nil, err
	}
	samples, err := s.Store.Samples.ListByBatch(id)
	if err != nil {
		return nil, err
	}
	valid := 0
	for _, smp := range samples {
		n, err := sample.ValidMotifsOfSample(s.Store.Motifs, smp.ID)
		if err != nil {
			return nil, err
		}
		valid += n
	}
	next, err := sample.NextBatchStatus(b.Status, len(samples), valid)
	if err != nil {
		return nil, err
	}
	if err := s.Store.Batches.UpdateStatus(id, next); err != nil {
		return nil, err
	}
	return s.Store.Batches.Get(id)
}

// SealBatch 封存批次（终态，禁止再录样本）。
func (s *Service) SealBatch(id int64) (*model.Batch, error) {
	b, err := s.Store.Batches.Get(id)
	if err != nil {
		return nil, err
	}
	if b.Status != model.BatchPublished {
		return nil, fmt.Errorf("%w: only published batch can be sealed", model.ErrState)
	}
	if err := s.Store.Batches.UpdateStatus(id, model.BatchSealed); err != nil {
		return nil, err
	}
	return s.Store.Batches.Get(id)
}

// ---- 样本 ----

// ImportSample 导入样本（幂等）。
func (s *Service) ImportSample(in sample.IngestInput) (*model.FabricSample, bool, error) {
	return s.Ingest.Ingest(in)
}

// ListSamples 列出样本（可按批次过滤）。
func (s *Service) ListSamples(batchID int64) ([]*model.FabricSample, error) {
	if batchID > 0 {
		return s.Store.Samples.ListByBatch(batchID)
	}
	return s.Store.Samples.List()
}

// GetSample 查询样本详情。
func (s *Service) GetSample(id int64) (*model.FabricSample, error) {
	return s.Store.Samples.Get(id)
}

// ---- 单元 ----

// AddMotif 添加纹样单元。
func (s *Service) AddMotif(sampleID int64, name string, originX, originY, width, height int, grid string) (*model.MotifUnit, error) {
	return sample.AddMotif(s.Store.Samples, s.Store.Motifs, sampleID,
		name, originX, originY, width, height, grid)
}

// ParseMotif 解析单元。
func (s *Service) ParseMotif(motifID int64) (*model.MotifUnit, error) {
	return sample.ParseMotif(s.Store.Motifs, motifID)
}

// ExcludeMotif 排除单元（补片/排除）。
func (s *Service) ExcludeMotif(motifID int64, asPatch bool) (*model.MotifUnit, error) {
	return sample.ExcludeMotif(s.Store.Motifs, motifID, asPatch)
}

// ListMotifs 列出单元（可按样本过滤）。
func (s *Service) ListMotifs(sampleID int64) ([]*model.MotifUnit, error) {
	if sampleID > 0 {
		return s.Store.Motifs.ListBySample(sampleID)
	}
	return s.Store.Motifs.List()
}

// GetMotif 查询单元详情。
func (s *Service) GetMotif(id int64) (*model.MotifUnit, error) {
	return s.Store.Motifs.Get(id)
}

// ---- 工艺特征 ----

// UpsertTechnique 写入工艺特征。
func (s *Service) UpsertTechnique(f *model.TechniqueFeature) (*model.TechniqueFeature, error) {
	if err := model.ValidateTechnique(f.WeaveClass, f.Interlacing, f.DyeClass, f.Colorfastness); err != nil {
		return nil, err
	}
	if _, err := s.Store.Samples.Get(f.SampleID); err != nil {
		return nil, err
	}
	return s.Store.Techniques.Upsert(f)
}

// GetTechnique 查询工艺特征。
func (s *Service) GetTechnique(sampleID int64) (*model.TechniqueFeature, error) {
	return s.Store.Techniques.BySample(sampleID)
}

// ListTechniques 列出全部工艺特征。
func (s *Service) ListTechniques() ([]*model.TechniqueFeature, error) {
	return s.Store.Techniques.List()
}

// VerifyTechnique 核验样本织法自洽性。
func (s *Service) VerifyTechnique(sampleID int64) (bool, string, error) {
	f, err := s.Store.Techniques.BySample(sampleID)
	if err != nil {
		return false, "", err
	}
	smp, err := s.Store.Samples.Get(sampleID)
	if err != nil {
		return false, "", err
	}
	ok, reason := technique.VerifyWeave(f, smp.WarpDensity, smp.WeftDensity)
	return ok, reason, nil
}

// ---- 关系候选 ----

// CompareMotifs 比较两个单元并生成/返回候选。
func (s *Service) CompareMotifs(fromID, toID int64) (*model.RelationCandidate, error) {
	from, err := s.Store.Motifs.Get(fromID)
	if err != nil {
		return nil, err
	}
	to, err := s.Store.Motifs.Get(toID)
	if err != nil {
		return nil, err
	}
	if from.SampleID == to.SampleID {
		return nil, fmt.Errorf("%w: cannot compare motifs of the same sample", model.ErrInvalid)
	}
	fromTech, _ := s.Store.Techniques.BySample(from.SampleID)
	toTech, _ := s.Store.Techniques.BySample(to.SampleID)

	gen, err := relation.Generate(relation.GenerateInput{
		From: from, To: to, FromTech: fromTech, ToTech: toTech,
	})
	if err != nil {
		return nil, err
	}
	// 已存在则返回既有候选（(from,to) 唯一），否则创建。
	existing, gerr := s.Store.Relations.GetByUnits(fromID, toID)
	if gerr == nil {
		return existing, nil
	}
	if gerr != model.ErrNotFound {
		return nil, gerr
	}
	return s.Store.Relations.Create(&model.RelationCandidate{
		FromUnitID:     fromID,
		ToUnitID:       toID,
		TopoSimilarity: gen.TopoSimilarity,
		WeaveCompat:    gen.WeaveCompat,
		DyeCompat:      gen.DyeCompat,
		Verdict:        gen.Verdict,
		Summary:        gen.Summary,
	})
}

// ListRelations 列出候选（可按裁决过滤）。
func (s *Service) ListRelations(verdict string) ([]*model.RelationCandidate, error) {
	return s.Store.Relations.List(verdict)
}

// GetRelation 查询候选详情（含反证）。
func (s *Service) GetRelation(id int64) (*model.RelationCandidate, []*model.CounterEvidence, error) {
	r, err := s.Store.Relations.Get(id)
	if err != nil {
		return nil, nil, err
	}
	list, err := s.Store.Evidence.ListByRelation(id)
	if err != nil {
		return nil, nil, err
	}
	return r, list, nil
}

// ApplyVerdict 人工裁决候选（版本校验防并发覆盖）。
func (s *Service) ApplyVerdict(id int64, verdict, summary string) (*model.RelationCandidate, error) {
	r, err := s.Store.Relations.Get(id)
	if err != nil {
		return nil, err
	}
	if err := relation.ApplyVerdict(r, verdict, summary); err != nil {
		return nil, err
	}
	if summary == "" {
		summary = r.Summary
	}
	if err := s.Store.Relations.UpdateVerdict(id, r.Verdict, verdict, summary); err != nil {
		return nil, err
	}
	return s.Store.Relations.Get(id)
}

// ResolveConflictVerdict 复核冲突候选（端到端场景）。
func (s *Service) ResolveConflictVerdict(id int64, oppose bool) (*model.RelationCandidate, error) {
	r, err := s.Store.Relations.Get(id)
	if err != nil {
		return nil, err
	}
	if r.Verdict != model.VerdictConflict {
		return nil, fmt.Errorf("%w: relation %d is not in conflict", model.ErrState, id)
	}
	next, summary := relation.ResolveConflict(r, oppose)
	if err := s.Store.Relations.UpdateVerdict(id, model.VerdictConflict, next, summary); err != nil {
		return nil, err
	}
	return s.Store.Relations.Get(id)
}

// ---- 反证 ----

// AddCounterEvidence 追加反证并自动评估对候选的影响。
func (s *Service) AddCounterEvidence(relationID int64, kind, description, ref string) (*model.CounterEvidence, error) {
	if err := evidence.ValidateKind(kind); err != nil {
		return nil, err
	}
	if _, err := s.Store.Relations.Get(relationID); err != nil {
		return nil, err
	}
	ev := &model.CounterEvidence{
		RelationID: relationID, Kind: kind, Description: description, Ref: ref,
	}
	created, err := s.Store.Evidence.Create(ev)
	if err != nil {
		return nil, err
	}
	// 自动按反证影响调整候选（出处/层位反证 → 否决；补片 → 降级）。
	r, err := s.Store.Relations.Get(relationID)
	if err != nil {
		return created, nil
	}
	if r.VersionID > 0 {
		// 已被版本收录的关系不再自动改写裁决，交由人工复核。
		return created, nil
	}
	items, err := s.Store.Evidence.ListByRelation(relationID)
	if err != nil {
		return created, nil
	}
	imp := evidence.AssessImpact(items, r.Verdict)
	if imp.Suggested != r.Verdict {
		_ = s.Store.Relations.UpdateVerdict(relationID, r.Verdict, imp.Suggested,
			imp.Explanation+"（自动）")
	}
	return created, nil
}

// ListEvidence 列出关系的反证。
func (s *Service) ListEvidence(relationID int64) ([]*model.CounterEvidence, error) {
	return s.Store.Evidence.ListByRelation(relationID)
}

// ---- 版本发布 ----

// CreateVersion 创建关系版本草稿：收录全部已确认/部分支持的关系，
// 发布前执行出处循环检测。
func (s *Service) CreateVersion(name, summary string) (*model.RelationVersion, error) {
	relations, err := s.Store.Relations.List("")
	if err != nil {
		return nil, err
	}
	var ids []string
	var edges [][2]int64
	for _, r := range relations {
		if r.Verdict == model.VerdictConfirmed || r.Verdict == model.VerdictPartial {
			ids = append(ids, fmt.Sprintf("%d", r.ID))
			edges = append(edges, [2]int64{r.FromUnitID, r.ToUnitID})
		}
	}
	if err := versioning.DetectCycle(edges); err != nil {
		return nil, err
	}
	if summary == "" {
		summary = fmt.Sprintf("收录 %d 条工艺传承关系", len(ids))
	}
	return s.Store.Versions.Create(name, summary, strings.Join(ids, ","), "")
}

// ShareVersion 共享版本（草稿 → shared）。
func (s *Service) ShareVersion(id int64) (*model.RelationVersion, error) {
	v, err := s.Store.Versions.Get(id)
	if err != nil {
		return nil, err
	}
	if err := versioning.Transition(v.Status, model.VersionShared); err != nil {
		return nil, err
	}
	// 共享时锁定关系归属。
	for _, rid := range store.ParseRelationIDs(v.RelationIDs) {
		if err := s.Store.Relations.AssignVersion(rid, v.ID); err != nil {
			return nil, err
		}
	}
	if err := s.Store.Versions.UpdateStatus(id, model.VersionShared, ""); err != nil {
		return nil, err
	}
	return s.Store.Versions.Get(id)
}

// FreezeVersion 冻结版本（shared → frozen）：保留出处与反证摘要快照。
func (s *Service) FreezeVersion(id int64) (*model.RelationVersion, error) {
	v, err := s.Store.Versions.Get(id)
	if err != nil {
		return nil, err
	}
	if err := versioning.Transition(v.Status, model.VersionFrozen); err != nil {
		return nil, err
	}
	// 汇总反证快照写入出处说明。
	rids := store.ParseRelationIDs(v.RelationIDs)
	var evs []*model.CounterEvidence
	for _, rid := range rids {
		list, err := s.Store.Evidence.ListByRelation(rid)
		if err != nil {
			return nil, err
		}
		evs = append(evs, list...)
	}
	note := evidence.ProvenanceNote(evs)
	if v.ProvenanceNote == "" {
		if err := s.setProvenanceNote(id, note); err != nil {
			return nil, err
		}
	}
	if err := s.Store.Versions.UpdateStatus(id, model.VersionFrozen, nowStr()); err != nil {
		return nil, err
	}
	return s.Store.Versions.Get(id)
}

// setProvenanceNote 写入版本出处说明。
func (s *Service) setProvenanceNote(id int64, note string) error {
	_, err := s.Store.DB.Exec(
		`UPDATE relation_versions SET provenance_note=? WHERE id=?`, note, id)
	return err
}

// SupersedeVersion 用新版本替代冻结版本。
func (s *Service) SupersedeVersion(id int64, newName, summary string) (*model.RelationVersion, error) {
	v, err := s.Store.Versions.Get(id)
	if err != nil {
		return nil, err
	}
	if err := versioning.Transition(v.Status, model.VersionSuperseded); err != nil {
		return nil, err
	}
	created, err := s.Store.Versions.Create(newName, summary, "", "替代版本 " + v.Name)
	if err != nil {
		return nil, err
	}
	if err := s.Store.Versions.MarkSuperseded(id, created.ID); err != nil {
		return nil, err
	}
	return s.Store.Versions.Get(id)
}

// ListVersions 列出全部版本。
func (s *Service) ListVersions() ([]*model.RelationVersion, error) {
	return s.Store.Versions.List()
}

// GetVersion 查询版本详情（含关系列表）。
func (s *Service) GetVersion(id int64) (*model.RelationVersion, []*model.RelationCandidate, error) {
	v, err := s.Store.Versions.Get(id)
	if err != nil {
		return nil, nil, err
	}
	rels, err := s.Store.Relations.ListByVersion(id)
	if err != nil {
		return nil, nil, err
	}
	return v, rels, nil
}

// ---- 统计与自检 ----

// Stats 汇总统计。
type Stats struct {
	Batches       int            `json:"batches"`
	Samples       int            `json:"samples"`
	Motifs        int            `json:"motifs"`
	ValidMotifs   int            `json:"valid_motifs"`
	Relations     map[string]int `json:"relations_by_verdict"`
	Versions      int            `json:"versions"`
	FrozenVersion int            `json:"frozen_versions"`
}

// GetStats 计算汇总统计。
func (s *Service) GetStats() (*Stats, error) {
	batches, err := s.Store.Batches.List()
	if err != nil {
		return nil, err
	}
	samples, err := s.Store.Samples.List()
	if err != nil {
		return nil, err
	}
	motifs, err := s.Store.Motifs.List()
	if err != nil {
		return nil, err
	}
	valid := 0
	for _, m := range motifs {
		if m.Status == model.MotifValid {
			valid++
		}
	}
	relStats, err := s.Store.Relations.Stats()
	if err != nil {
		return nil, err
	}
	versions, err := s.Store.Versions.List()
	if err != nil {
		return nil, err
	}
	frozen := 0
	for _, v := range versions {
		if v.Status == model.VersionFrozen {
			frozen++
		}
	}
	return &Stats{
		Batches:       len(batches),
		Samples:       len(samples),
		Motifs:        len(motifs),
		ValidMotifs:   valid,
		Relations:     relStats,
		Versions:      len(versions),
		FrozenVersion: frozen,
	}, nil
}

// SelfCheck 自检：验证数据库可读写、实体可查询、工艺验证函数可用。
func (s *Service) SelfCheck() (map[string]string, error) {
	out := map[string]string{}
	if err := s.Store.DB.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	out["db"] = "ok"
	batches, err := s.Store.Batches.List()
	if err != nil {
		return nil, err
	}
	out["batches"] = fmt.Sprintf("%d", len(batches))
	rels, err := s.Store.Relations.List("")
	if err != nil {
		return nil, err
	}
	out["relations"] = fmt.Sprintf("%d", len(rels))
	out["topology"] = "ok"
	return out, nil
}

// SortRelationsBySimilarity 按拓扑相似度降序排序（统计/展示用）。
func SortRelationsBySimilarity(rels []*model.RelationCandidate) {
	sort.SliceStable(rels, func(i, j int) bool {
		return rels[i].TopoSimilarity > rels[j].TopoSimilarity
	})
}

func nowStr() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
}
