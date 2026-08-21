FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /dogcam-server ./dogcam/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /dogcam-server /dogcam-server
EXPOSE 8080 50051 9090
ENTRYPOINT ["/dogcam-server"]
