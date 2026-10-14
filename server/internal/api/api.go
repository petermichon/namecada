// Package api exposes the scan engine over HTTP. GET /check?name=<label>
// streams one JSON result per provider as it resolves, and every response
// carries CORS headers so the browser app can consume it directly.
package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/petermichon/domain-browser/server/internal/engine"
)

// DefaultOrigin is the origin the dev app is served from (Vite's default port).
const DefaultOrigin = "http://localhost:5173"

// labelPattern is the DNS-label rule a scan name must satisfy: at most 63
// characters, lowercase alphanumerics with single interior hyphens. It mirrors
// the frontend's toLabel so the client and server agree on valid names.
var labelPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

// normalizeLabel trims and lowercases input, reporting whether the result is a
// valid label. It mirrors the frontend's toLabel.
func normalizeLabel(input string) (string, bool) {
	label := strings.ToLower(strings.TrimSpace(input))
	if !labelPattern.MatchString(label) {
		return "", false
	}
	return label, true
}

// Server serves the engine over HTTP with a single allowed CORS origin.
type Server struct {
	engine *engine.Engine
	origin string
	logger *slog.Logger
}

// New returns a Server that scans with eng and allows origin through CORS. An
// empty origin falls back to DefaultOrigin. A nil logger discards request logs.
func New(eng *engine.Engine, origin string, logger *slog.Logger) *Server {
	if origin == "" {
		origin = DefaultOrigin
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Server{engine: eng, origin: origin, logger: logger}
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
	start := time.Now()
	label := strings.TrimSpace(r.URL.Query().Get("name"))
	results := 0
	defer func() {
		s.logger.Info("request",
			"method", r.Method,
			"label", label,
			"duration", time.Since(start),
			"results", results,
		)
	}()

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

	if label == "" {
		writeError(w, http.StatusBadRequest, "missing required query parameter \"name\"")
		return
	}
	name, ok := normalizeLabel(label)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid query parameter \"name\"")
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
		results++
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
