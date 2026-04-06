# Stage 1: build Go binary
# Context: repo root (docker build -f capi-argocd-bootstrapper/Dockerfile .)
FROM golang:1.26-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/capi-argocd-bootstrapper ./capi-argocd-bootstrapper/

# Stage 2: minimal runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=go-builder /out/capi-argocd-bootstrapper /usr/local/bin/capi-argocd-bootstrapper
ENTRYPOINT ["/usr/local/bin/capi-argocd-bootstrapper"]
