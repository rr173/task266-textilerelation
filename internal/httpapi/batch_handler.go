package httpapi

import (
	"net/http"
)

// createBatch POST /api/batches
func (h *Handler) createBatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Note string `json:"note"`
	}
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	b, err := h.svc.CreateBatch(req.Name, req.Note)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

// listBatches GET /api/batches
func (h *Handler) listBatches(w http.ResponseWriter, _ *http.Request) {
	list, err := h.svc.ListBatches()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// getBatch GET /api/batches/{id}
func (h *Handler) getBatch(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	b, err := h.svc.GetBatch(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// advanceBatch POST /api/batches/{id}/advance
func (h *Handler) advanceBatch(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	b, err := h.svc.AdvanceBatch(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// sealBatch POST /api/batches/{id}/seal
func (h *Handler) sealBatch(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	b, err := h.svc.SealBatch(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}
