package httpapi

import (
	"net/http"
)

// createVersion POST /api/versions
func (h *Handler) createVersion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Summary string `json:"summary"`
	}
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	v, err := h.svc.CreateVersion(req.Name, req.Summary)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

// listVersions GET /api/versions
func (h *Handler) listVersions(w http.ResponseWriter, _ *http.Request) {
	list, err := h.svc.ListVersions()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// getVersion GET /api/versions/{id}
func (h *Handler) getVersion(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	v, rels, err := h.svc.GetVersion(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"version":   v,
		"relations": rels,
	})
}

// shareVersion POST /api/versions/{id}/share
func (h *Handler) shareVersion(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	v, err := h.svc.ShareVersion(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// freezeVersion POST /api/versions/{id}/freeze
func (h *Handler) freezeVersion(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	v, err := h.svc.FreezeVersion(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// supersedeVersion POST /api/versions/{id}/supersede
func (h *Handler) supersedeVersion(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		Name    string `json:"name"`
		Summary string `json:"summary"`
	}
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	v, err := h.svc.SupersedeVersion(id, req.Name, req.Summary)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}
