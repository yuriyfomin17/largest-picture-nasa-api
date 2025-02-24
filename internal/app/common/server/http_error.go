package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func BadRequest(slug string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, slug, w, r, "Bad request", http.StatusBadRequest)
}

func NotFound(slug string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, slug, w, r, "Not found", http.StatusNotFound)
}

func httpRespondWithError(err error, slug string, w http.ResponseWriter, r *http.Request, msg string, status int) {
	log.Printf("error: %s, slug: %s, msg: %s", err, slug, msg)

	resp := ErrorResponse{Slug: slug, httpStatus: status}
	if os.Getenv("DEBUG_ERRORS") != "" && err != nil {
		resp.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

type ErrorResponse struct {
	Slug       string `json:"slug"`
	Error      string `json:"error,omitempty"`
	httpStatus int
}

func (e ErrorResponse) Render(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(e.httpStatus)
	return nil
}
