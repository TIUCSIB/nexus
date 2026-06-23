FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o nexus ./cmd/nexus

FROM alpine:latest
RUN apk add --no-cache ca-certificates sqlite
WORKDIR /app
COPY --from=builder /app/nexus .
COPY config.yaml .
EXPOSE 8080 9090
CMD ["./nexus"]