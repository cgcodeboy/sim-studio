// Package respond 写 HTTP 响应。各业务 handler 共用它，不互相调用。
package respond

import (
	"encoding/json"
	"errors"
	"net/http"

	model "github.com/cgcodeboy/sim-studio/internal/pkgs/model/domain"
	record "github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
)

func Decode(w http.ResponseWriter, r *http.Request, dest any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20+4096)
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": model.CodeInvalid})
		return err
	}
	return nil
}

func Fail(w http.ResponseWriter, err error) {
	var de *model.Error
	if errors.As(err, &de) {
		writeCoded(w, de.Code, de.Project)
		return
	}
	if re := record.As(err); re != nil {
		writeCoded(w, re.Code, nil)
		return
	}
	WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"error": model.CodeUnavailable})
}

func writeCoded(w http.ResponseWriter, code string, project *model.Project) {
	status := http.StatusServiceUnavailable
	switch code {
	case model.CodeInvalid:
		status = http.StatusUnprocessableEntity
	case model.CodeNotFound:
		status = http.StatusNotFound
	case model.CodeConflict, model.CodeRejected:
		status = http.StatusConflict
	}
	body := map[string]any{"error": code}
	if project != nil {
		body["project"] = project
	}
	WriteJSON(w, status, body)
}

func WriteJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
