package metricsstore

import (
	"encoding/json"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
)

// Handler returns an http.Handler that serves the latest MetricsPayload as JSON.
// No auth — intended for internal/metrics port only.
func Handler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := s.Get()
		if p == nil {
			_ = json.NewEncoder(w).Encode(nil)
			return
		}
		b, err := protojson.Marshal(p)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(b)
	})
}
