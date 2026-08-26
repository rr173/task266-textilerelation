package httpapi

import (
	"net/http"
)

// health GET /api/health
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// stats GET /api/stats
func (h *Handler) stats(w http.ResponseWriter, _ *http.Request) {
	st, err := h.svc.GetStats()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// selfCheck GET /api/selfcheck
func (h *Handler) selfCheck(w http.ResponseWriter, _ *http.Request) {
	out, err := h.svc.SelfCheck()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
