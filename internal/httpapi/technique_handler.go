package httpapi

import (
	"net/http"

	"task266-textilerelation/internal/model"
)

// upsertTechnique PUT /api/samples/{id}/technique
func (h *Handler) upsertTechnique(w http.ResponseWriter, r *http.Request) {
	sampleID, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		WeaveClass     string  `json:"weave_class"`
		Interlacing    string  `json:"interlacing"`
		TwillDirection string  `json:"twill_direction"`
		YarnTwist      string  `json:"yarn_twist"`
		DyeClass       string  `json:"dye_class"`
		DyePigment     string  `json:"dye_pigment"`
		Colorfastness  int     `json:"colorfastness"`
		CarbonRatio    float64 `json:"carbon_ratio"`
	}
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	f, err := h.svc.UpsertTechnique(&model.TechniqueFeature{
		SampleID:       sampleID,
		WeaveClass:     req.WeaveClass,
		Interlacing:    req.Interlacing,
		TwillDirection: req.TwillDirection,
		YarnTwist:      req.YarnTwist,
		DyeClass:       req.DyeClass,
		DyePigment:     req.DyePigment,
		Colorfastness:  req.Colorfastness,
		CarbonRatio:    req.CarbonRatio,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

// getTechnique GET /api/samples/{id}/technique
func (h *Handler) getTechnique(w http.ResponseWriter, r *http.Request) {
	sampleID, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	f, err := h.svc.GetTechnique(sampleID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

// listTechniques GET /api/techniques
func (h *Handler) listTechniques(w http.ResponseWriter, _ *http.Request) {
	list, err := h.svc.ListTechniques()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// verifyTechnique POST /api/samples/{id}/verify-technique
func (h *Handler) verifyTechnique(w http.ResponseWriter, r *http.Request) {
	sampleID, err := pathID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	ok, reason, err := h.svc.VerifyTechnique(sampleID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sample_id": sampleID,
		"ok":        ok,
		"reason":    reason,
	})
}
