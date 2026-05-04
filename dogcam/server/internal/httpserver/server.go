package httpserver

import (
	"embed"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/fergalhk-lab/apps/dogcam/server/internal/broadcast"
	"github.com/google/uuid"
)

//go:embed static
var staticFiles embed.FS

type Server struct {
	handler http.Handler
}

func New(b *broadcast.Broadcaster, viewerPassword string) *Server {
	mux := http.NewServeMux()

	static, _ := fs.Sub(staticFiles, "static")
	mux.Handle("GET /", http.FileServerFS(static))
	mux.Handle("GET /stream", sseHandler(b))

	return &Server{handler: basicAuth(viewerPassword, mux)}
}

func (s *Server) Handler() http.Handler {
	return s.handler
}

func basicAuth(password string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pass, ok := r.BasicAuth()
		if !ok || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="dogcam"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func sseHandler(b *broadcast.Broadcaster) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		id := uuid.New().String()
		frames := b.Subscribe(id)
		defer b.Unsubscribe(id)

		for {
			select {
			case frame, ok := <-frames:
				if !ok {
					return
				}
				fmt.Fprintf(w, "event: frame\ndata: %s\n\n", base64.StdEncoding.EncodeToString(frame))
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	})
}
