// Package httpapi 提供 /api 前缀的 HTTP 路由与处理器。
package httpapi

import (
	"net/http"

	"task266-textilerelation/internal/service"
)

// Handler 聚合 service 与路由。
type Handler struct {
	svc *service.Service
	mux *http.ServeMux
}

// New 构造 HTTP 处理器。
func New(svc *service.Service) *Handler {
	h := &Handler{svc: svc, mux: http.NewServeMux()}
	h.routes()
	return h
}

// Handler 返回 http.Handler（实现 http.Handler 接口）。
func (h *Handler) Handler() http.Handler {
	return h.mux
}

// routes 注册全部路由（统一前缀 /api）。
func (h *Handler) routes() {
	// 批次
	h.mux.HandleFunc("POST /api/batches", h.createBatch)
	h.mux.HandleFunc("GET /api/batches", h.listBatches)
	h.mux.HandleFunc("GET /api/batches/{id}", h.getBatch)
	h.mux.HandleFunc("POST /api/batches/{id}/advance", h.advanceBatch)
	h.mux.HandleFunc("POST /api/batches/{id}/seal", h.sealBatch)

	// 样本
	h.mux.HandleFunc("POST /api/samples", h.importSample)
	h.mux.HandleFunc("GET /api/samples", h.listSamples)
	h.mux.HandleFunc("GET /api/samples/{id}", h.getSample)

	// 纹样单元
	h.mux.HandleFunc("POST /api/samples/{id}/motifs", h.addMotif)
	h.mux.HandleFunc("GET /api/samples/{id}/motifs", h.listSampleMotifs)
	h.mux.HandleFunc("GET /api/motifs", h.listMotifs)
	h.mux.HandleFunc("GET /api/motifs/{id}", h.getMotif)
	h.mux.HandleFunc("POST /api/motifs/{id}/parse", h.parseMotif)
	h.mux.HandleFunc("POST /api/motifs/{id}/exclude", h.excludeMotif)

	// 工艺特征
	h.mux.HandleFunc("PUT /api/samples/{id}/technique", h.upsertTechnique)
	h.mux.HandleFunc("GET /api/samples/{id}/technique", h.getTechnique)
	h.mux.HandleFunc("GET /api/techniques", h.listTechniques)
	h.mux.HandleFunc("POST /api/samples/{id}/verify-technique", h.verifyTechnique)

	// 关系候选
	h.mux.HandleFunc("POST /api/motifs/{from}/compare/{to}", h.compareMotifs)
	h.mux.HandleFunc("GET /api/relations", h.listRelations)
	h.mux.HandleFunc("GET /api/relations/{id}", h.getRelation)
	h.mux.HandleFunc("POST /api/relations/{id}/verdict", h.applyVerdict)
	h.mux.HandleFunc("POST /api/relations/{id}/resolve-conflict", h.resolveConflict)

	// 反证
	h.mux.HandleFunc("POST /api/relations/{id}/evidence", h.addEvidence)
	h.mux.HandleFunc("GET /api/relations/{id}/evidence", h.listEvidence)

	// 版本发布
	h.mux.HandleFunc("POST /api/versions", h.createVersion)
	h.mux.HandleFunc("GET /api/versions", h.listVersions)
	h.mux.HandleFunc("GET /api/versions/{id}", h.getVersion)
	h.mux.HandleFunc("POST /api/versions/{id}/share", h.shareVersion)
	h.mux.HandleFunc("POST /api/versions/{id}/freeze", h.freezeVersion)
	h.mux.HandleFunc("POST /api/versions/{id}/supersede", h.supersedeVersion)

	// 运维
	h.mux.HandleFunc("GET /api/health", h.health)
	h.mux.HandleFunc("GET /api/stats", h.stats)
	h.mux.HandleFunc("GET /api/selfcheck", h.selfCheck)
}
