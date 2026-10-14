// Package api exposes the scan engine over HTTP. GET /check?name=<label>
// streams one JSON result per provider as it resolves, and every response
// carries CORS headers so the browser app can consume it directly.
package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/petermichon/domain-browser/backend/internal/engine"
)

// DefaultOrigin is the origin the dev app is served from (Vite's default port).
const DefaultOrigin = "http://localhost:5173"

// Server serves the engine over HTTP with a single allowed CORS origin.
type Server struct {
	engine *engine.Engine
	origin string
}

// New returns a Server that scans with eng and allows origin through CORS. An
// empty origin falls back to DefaultOrigin.
func New(eng *engine.Engine, origin string) *Server {
	if origin == "" {
		origin = DefaultOrigin
	}
	return &Server{engine: eng, origin: origin}
}

// Handler returns the HTTP routes for the server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/check", s.handleCheck)
	return mux
}

// handleCheck streams one NDJSON result per provider for ?name=<label>. A
// fresh in-memory cache is reused across requests by the engine, so repeated
// labels are served without re-checking the sources.
func (s *Server) handleCheck(w http.ResponseWriter, r *http.Request) {
	s.setCORS(w)

	switch r.Method {
	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodGet:
	default:
		w.Header().Set("Allow", "GET, OPTIONS")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		writeError(w, http.StatusBadRequest, "missing required query parameter \"name\"")
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	flusher, _ := w.(http.Flusher)
	enc := json.NewEncoder(w)
	for res := range s.engine.Scan(r.Context(), name) {
		if err := enc.Encode(res); err != nil {
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
	}
}

// setCORS writes the headers that let the app origin read the response.
func (s *Server) setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", s.origin)
	w.Header().Add("Vary", "Origin")
}

// writeError sends a small JSON error body, keeping responses JSON even on
// failure so the client can parse them uniformly.
func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
