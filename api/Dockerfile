# syntax=docker/dockerfile:1

FROM golang:1.26.5-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && \
    addgroup -S memoria && adduser -S memoria -G memoria
WORKDIR /app

COPY --from=builder /out/api ./api
COPY --from=builder /src/docs ./docs

USER memoria
EXPOSE 3000
ENTRYPOINT ["./api"]
