FROM golang:1.25-alpine AS builder

ARG VERSION=dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go build -ldflags "-X main.version=${VERSION}" -o /out/control-plane-api ./cmd/control-plane-api

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /out/control-plane-api /usr/local/bin/control-plane-api
EXPOSE 8080
ENTRYPOINT ["control-plane-api"]
