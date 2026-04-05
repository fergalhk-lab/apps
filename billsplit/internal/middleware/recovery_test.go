// billsplit/internal/middleware/recovery_test.go
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

func TestRecoverPanic_returns500OnPanic(t *testing.T) {
	core, logs := observer.New(zapcore.ErrorLevel)
	logger := zap.New(core)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()
	middleware.RecoverPanic(logger)(inner).ServeHTTP(rw, req)

	if rw.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rw.Code)
	}
	if logs.Len() != 1 {
		t.Fatalf("expected 1 error log, got %d", logs.Len())
	}
}

func TestRecoverPanic_passesThrough(t *testing.T) {
	logger := zap.NewNop()

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()
	middleware.RecoverPanic(logger)(inner).ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rw.Code)
	}
}
