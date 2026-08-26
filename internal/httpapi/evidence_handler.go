package httpapi

import (
	"net/http"
)

// addEvidence POST /api/relations/{id}/evidence
func (h *Handler) addEvidence(w http.ResponseWriter, r *http.Request) {
	relationID, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		Kind        string `json:"kind"`
		Description string `json:"description"`
		Ref         string `json:"ref"`
	}
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	ev, err := h.svc.AddCounterEvidence(relationID, req.Kind, req.Description, req.Ref)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ev)
}

// listEvidence GET /api/relations/{id}/evidence
func (h *Handler) listEvidence(w http.ResponseWriter, r *http.Request) {
	relationID, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	list, err := h.svc.ListEvidence(relationID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}
