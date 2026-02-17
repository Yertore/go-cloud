# ---------- build stage ----------
FROM golang:1.25 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/app ./


# ---------- runtime stage ----------
FROM alpine:3.20

ENV PORT=8080

WORKDIR /

# Чтобы TLS/https запросы работали (если ты будешь ходить во внешние API)
RUN apk add --no-cache ca-certificates

COPY --from=builder /out/app /app

EXPOSE 8080

ENTRYPOINT ["/app"]
