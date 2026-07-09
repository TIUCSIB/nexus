# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app
COPY web/package.json web/package-lock.json* ./
RUN npm ci --omit=optional
COPY web/ .
RUN npm run build

# Stage 2: Build panel
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o /app/nexus ./cmd/nexus
RUN CGO_ENABLED=1 go build -o /app/nexus-agent ./agent/cmd/agent

# Stage 3: Runtime image
FROM alpine:3.21
RUN apk add --no-cache ca-certificates sqlite wget bash
WORKDIR /app

COPY --from=builder /app/nexus .
COPY --from=builder /app/nexus-agent .
COPY --from=frontend /app/dist ./web/dist
COPY config.yaml .

# Create entrypoint script for first-run setup
RUN echo '#!/bin/bash
set -e
if [ ! -f /app/data/nexus.db ]; then
  echo "First run: initializing database..."
  mkdir -p /app/data
fi
echo "Starting Nexus panel..."
exec ./nexus
' > /app/entrypoint.sh && chmod +x /app/entrypoint.sh

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]