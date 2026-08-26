package httpapi

import (
	"net/http"
	"strconv"

	"task266-textilerelation/internal/sample"
)

// importSample POST /api/samples
func (h *Handler) importSample(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BatchID     int64             `json:"batch_id"`
		Name        string            `json:"name"`
		Provenance  string            `json:"provenance"`
		WarpCount   int               `json:"warp_count"`
		WeftCount   int               `json:"weft_count"`
		WarpDensity int               `json:"warp_density"`
		WeftDensity int               `json:"weft_density"`
		MotifGrids  map[string]string `json:"motif_grids"`
	}
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	sm, created, err := h.svc.ImportSample(sample.IngestInput{
		BatchID:     req.BatchID,
		Name:        req.Name,
		Provenance:  req.Provenance,
		WarpCount:   req.WarpCount,
		WeftCount:   req.WeftCount,
		WarpDensity: req.WarpDensity,
		WeftDensity: req.WeftDensity,
		MotifGrids:  req.MotifGrids,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK // 幂等命中：返回既有样本
	}
	writeJSON(w, status, sm)
}

// listSamples GET /api/samples?batch_id=
func (h *Handler) listSamples(w http.ResponseWriter, r *http.Request) {
	batchID := int64(0)
	if v := r.URL.Query().Get("batch_id"); v != "" {
		id, err := parseID(v)
		if err != nil {
			writeErr(w, err)
			return
		}
		batchID = id
	}
	list, err := h.svc.ListSamples(batchID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// getSample GET /api/samples/{id}
func (h *Handler) getSample(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	sm, err := h.svc.GetSample(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sm)
}

// parseID 解析查询参数中的 id。
func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
