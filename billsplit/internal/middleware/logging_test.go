// billsplit/internal/middleware/logging_test.go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fergalhk-lab/apps/billsplit/internal/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLogger_emitsOneLogPerRequest(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()
	middleware.RequestLogger(logger)(inner).ServeHTTP(rw, req)

	if logs.Len() != 1 {
		t.Fatalf("expected 1 log line, got %d", logs.Len())
	}
	if rw.Code != http.StatusCreated {
		t.Errorf("inner handler status not preserved: got %d, want 201", rw.Code)
	}
}

func TestRequestLogger_defaultStatus200WhenWriteHeaderNotCalled(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// deliberately omits WriteHeader
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	middleware.RequestLogger(logger)(inner).ServeHTTP(httptest.NewRecorder(), req)

	if logs.Len() != 1 {
		t.Fatalf("expected 1 log line, got %d", logs.Len())
	}
	entry := logs.All()[0]
	for _, f := range entry.Context {
		if f.Key == "status" && f.Integer != http.StatusOK {
			t.Errorf("expected logged status 200, got %d", f.Integer)
		}
	}
}
