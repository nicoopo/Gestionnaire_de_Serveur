# --- Étape de build ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server-manager ./cmd/server

# --- Étape finale (image légère) ---
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/server-manager .
COPY --from=builder /app/web ./web

EXPOSE 8080

CMD ["./server-manager"]