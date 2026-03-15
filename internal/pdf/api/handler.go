package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"apiservices/pdf-analysis/internal/pdf/analysis"
)

type Handler struct {
	service *analysis.Service
}

func NewHandler(service *analysis.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/v1/pdf/") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/v1/pdf/"), "/")
	switch path {
	case "analyze":
		h.handleAnalyze(w, r)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (h *Handler) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := r.ParseMultipartForm(h.service.MaxFileSize() + (1 << 20)); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, h.service.MaxFileSize()+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read file")
		return
	}
	if int64(len(data)) > h.service.MaxFileSize() {
		writeError(w, http.StatusBadRequest, "file too large")
		return
	}

	result, err := h.service.Analyze(r.Context(), data, header.Filename)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"failed to marshal response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}
