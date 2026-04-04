# Stage 1: build frontend
# Context: repo root (docker build -f billsplit/Dockerfile .)
FROM node:22-alpine AS frontend-builder
WORKDIR /app/billsplit/frontend
COPY billsplit/frontend/package*.json ./
RUN npm ci
COPY billsplit/frontend/ ./
RUN npm run build

# Stage 2: build Go binary
FROM golang:1.25-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/billsplit/frontend/dist ./billsplit/frontend/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/billsplit ./billsplit/ && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/seed ./billsplit/cmd/seed/

# Stage 3: minimal runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=go-builder /out/billsplit /usr/local/bin/billsplit
COPY --from=go-builder /out/seed /usr/local/bin/seed
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/billsplit"]
