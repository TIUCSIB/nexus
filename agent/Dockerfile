FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o nexus-agent ./cmd/agent

FROM alpine:latest
RUN apk add --no-cache ca-certificates sing-box
WORKDIR /app
COPY --from=builder /app/nexus-agent .
COPY agent.yaml .
CMD ["./nexus-agent"]