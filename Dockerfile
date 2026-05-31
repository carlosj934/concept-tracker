# ── Stage 1: Build ──────────────────────────────────────────────
FROM golang:1.26.3-alpine AS builder

WORKDIR /app

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN go build -trimpath -ldflags "-w -s" -o /app/bin/ct-api ./cmd/api

# ── Stage 2: Install migrate CLI ────────────────────────────────
FROM golang:1.26.3-alpine AS migrate-builder

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# ── Stage 3: Final image ─────────────────────────────────────────
FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy API binary
COPY --from=builder /app/bin/ct-api .

# Copy migrate binary
COPY --from=migrate-builder /go/bin/migrate /usr/local/bin/migrate

# Copy migrations
COPY db/migrations ./db/migrations

USER appuser

EXPOSE 8080

CMD ["./ct-api"]
