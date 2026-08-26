package httpapi

import (
	"net/http"
)

// compareMotifs POST /api/motifs/{from}/compare/{to}
func (h *Handler) compareMotifs(w http.ResponseWriter, r *http.Request) {
	fromRaw := r.PathValue("from")
	toRaw := r.PathValue("to")
	fromID, err := parseID(fromRaw)
	if err != nil {
		writeErr(w, err)
		return
	}
	toID, err := parseID(toRaw)
	if err != nil {
		writeErr(w, err)
		return
	}
	rel, err := h.svc.CompareMotifs(fromID, toID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rel)
}

// listRelations GET /api/relations?verdict=
func (h *Handler) listRelations(w http.ResponseWriter, r *http.Request) {
	verdict := r.URL.Query().Get("verdict")
	list, err := h.svc.ListRelations(verdict)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// getRelation GET /api/relations/{id}
func (h *Handler) getRelation(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	rel, evs, err := h.svc.GetRelation(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"relation": rel,
		"evidence": evs,
	})
}

// applyVerdict POST /api/relations/{id}/verdict
func (h *Handler) applyVerdict(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		Verdict string `json:"verdict"`
		Summary string `json:"summary"`
	}
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	rel, err := h.svc.ApplyVerdict(id, req.Verdict, req.Summary)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rel)
}

// resolveConflict POST /api/relations/{id}/resolve-conflict
func (h *Handler) resolveConflict(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		Oppose bool `json:"oppose"`
	}
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	rel, err := h.svc.ResolveConflictVerdict(id, req.Oppose)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rel)
}
