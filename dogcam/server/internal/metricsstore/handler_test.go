package metricsstore_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fergalhk-lab/apps/dogcam/gen/dogcampb"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/metricsstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsHandler_ReturnsNullWhenEmpty(t *testing.T) {
	s := metricsstore.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	metricsstore.Handler(s).ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `null`, w.Body.String())
}

func TestMetricsHandler_ReturnsLatestPayload(t *testing.T) {
	s := metricsstore.New()
	s.Update(&dogcampb.MetricsPayload{FramesCaptured: 5, FramesSent: 4, StreamingActive: true})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	metricsstore.Handler(s).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	// protojson encodes int64 fields as strings per the proto3 JSON spec
	assert.Equal(t, "5", got["framesCaptured"])
}
