package httpapi

import (
	"net/http"
)

// addMotif POST /api/samples/{id}/motifs
func (h *Handler) addMotif(w http.ResponseWriter, r *http.Request) {
	sampleID, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		Name     string `json:"name"`
		OriginX  int    `json:"origin_x"`
		OriginY  int    `json:"origin_y"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		Grid     string `json:"grid"`
	}
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	u, err := h.svc.AddMotif(sampleID, req.Name, req.OriginX, req.OriginY,
		req.Width, req.Height, req.Grid)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

// listSampleMotifs GET /api/samples/{id}/motifs
func (h *Handler) listSampleMotifs(w http.ResponseWriter, r *http.Request) {
	sampleID, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	list, err := h.svc.ListMotifs(sampleID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// listMotifs GET /api/motifs
func (h *Handler) listMotifs(w http.ResponseWriter, _ *http.Request) {
	list, err := h.svc.ListMotifs(0)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// getMotif GET /api/motifs/{id}
func (h *Handler) getMotif(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	u, err := h.svc.GetMotif(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// parseMotif POST /api/motifs/{id}/parse
func (h *Handler) parseMotif(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	u, err := h.svc.ParseMotif(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// excludeMotif POST /api/motifs/{id}/exclude?patch=true
func (h *Handler) excludeMotif(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	asPatch := r.URL.Query().Get("patch") == "true"
	u, err := h.svc.ExcludeMotif(id, asPatch)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}
