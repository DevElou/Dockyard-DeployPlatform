FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/control-plane-api ./cmd/control-plane-api

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /out/control-plane-api /usr/local/bin/control-plane-api
EXPOSE 8080
ENTRYPOINT ["control-plane-api"]
