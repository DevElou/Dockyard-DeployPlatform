FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/control-plane-api ./cmd/control-plane-api

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /out/control-plane-api /usr/local/bin/control-plane-api
EXPOSE 8080
ENTRYPOINT ["control-plane-api"]
