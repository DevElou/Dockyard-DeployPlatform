FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/deploy-agent ./cmd/deploy-agent

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /out/deploy-agent /usr/local/bin/deploy-agent
EXPOSE 8090
ENTRYPOINT ["deploy-agent"]
