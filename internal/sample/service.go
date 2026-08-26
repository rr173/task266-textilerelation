package sample

import (
	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/store"
)

// IngestService 是样本导入的服务门面，持有仓库引用以便编排批次、样本与单元。
type IngestService struct {
	store *store.Store
}

// NewIngestService 构造导入服务。
func NewIngestService(st *store.Store) *IngestService {
	return &IngestService{store: st}
}

// Ingest 委托包级函数执行样本导入（幂等）。
func (s *IngestService) Ingest(in IngestInput) (*model.FabricSample, bool, error) {
	return Ingest(s.store.Samples, s.store.Motifs, in)
}
