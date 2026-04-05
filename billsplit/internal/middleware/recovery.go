// billsplit/internal/middleware/recovery.go
package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

func RecoverPanic(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					logger.Error("panic recovered",
						zap.String("panic", fmt.Sprintf("%v", v)),
						zap.String("stack", string(debug.Stack())),
					)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
