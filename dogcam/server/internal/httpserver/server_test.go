package httpserver_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fergalhk-lab/apps/dogcam/server/internal/broadcast"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/httpserver"
	"github.com/fergalhk-lab/apps/dogcam/server/internal/metricsstore"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuth_RejectsNoCredentials(t *testing.T) {
	b := broadcast.New(2000)
	ms := metricsstore.New()
	srv := httpserver.New(b, ms, "testpass")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, `Basic realm="dogcam"`, w.Header().Get("WWW-Authenticate"))
}

func TestBasicAuth_RejectsWrongPassword(t *testing.T) {
	b := broadcast.New(2000)
	ms := metricsstore.New()
	srv := httpserver.New(b, ms, "testpass")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("viewer", "wrongpass")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestBasicAuth_AllowsCorrectPassword(t *testing.T) {
	b := broadcast.New(2000)
	ms := metricsstore.New()
	srv := httpserver.New(b, ms, "testpass")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("viewer", "testpass")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// flushRecorder supports http.Flusher for SSE testing.
type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed chan struct{}
}

func newFlushRecorder() *flushRecorder {
	return &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		flushed:          make(chan struct{}, 16),
	}
}

func (f *flushRecorder) Flush() {
	select {
	case f.flushed <- struct{}{}:
	default:
	}
}

func TestSSE_PublishesFrameAsBase64Event(t *testing.T) {
	b := broadcast.New(2000)
	ms := metricsstore.New()
	srv := httpserver.New(b, ms, "testpass")

	frame := []byte{0xFF, 0xD8, 0xFF}
	go func() {
		// Publish repeatedly until the SSE handler subscribes and receives
		for range [10]struct{}{} {
			b.Publish(frame)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/stream", nil).WithContext(ctx)
	req.SetBasicAuth("viewer", "testpass")
	w := newFlushRecorder()

	// Run handler in goroutine; cancel after first flush.
	done := make(chan struct{})
	go func() {
		defer close(done)
		srv.Handler().ServeHTTP(w, req)
	}()

	select {
	case <-w.flushed:
		cancel()
	case <-done:
	}
	<-done

	body := w.Body.String()
	encoded := base64.StdEncoding.EncodeToString(frame)
	assert.Contains(t, body, "event: frame\n")
	assert.Contains(t, body, "data: "+encoded)
}

