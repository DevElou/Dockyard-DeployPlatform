FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/orchestrator-worker ./cmd/orchestrator-worker

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /out/orchestrator-worker /usr/local/bin/orchestrator-worker
ENTRYPOINT ["orchestrator-worker"]
