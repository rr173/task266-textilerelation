package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/store"
)

// writeJSON 输出 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeErr 把领域错误映射为 HTTP 状态码与错误体。
func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, model.ErrConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrState):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrFrozen):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrCycle):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrTechnique):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	default:
		var sm *store.StateMismatchError
		if errors.As(err, &sm) {
			writeJSON(w, http.StatusPreconditionFailed, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal: " + err.Error()})
	}
}

// pathID 解析路径中的 int64 参数。
func pathID(r *http.Request) (int64, error) {
	raw := r.PathValue("id")
	if raw == "" {
		return 0, fmt.Errorf("%w: missing id", model.ErrInvalid)
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%w: bad id %q", model.ErrInvalid, raw)
	}
	return id, nil
}

// decodeBody 解析 JSON 请求体到目标结构。
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad json: " + err.Error()})
		return err
	}
	return nil
}
