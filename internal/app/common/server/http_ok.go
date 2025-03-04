package server

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Sol     int    `json:"sol,omitempty"`
	ImgSrc  string `json:"img_src,omitempty"`
	Size    int    `json:"size,omitempty"`
	Message string `json:"message,omitempty"`
}

func RespondOK(data SuccessResponse, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}
