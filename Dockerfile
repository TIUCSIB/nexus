# Nexus 多环境部署配置

FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o nexus ./cmd/nexus/

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata sqlite
WORKDIR /app
COPY --from=builder /app/nexus .
RUN mkdir -p /data /app/web/dist

# 前端文件需要外部挂载或 COPY（见部署文档）
EXPOSE 8080
CMD ["./nexus", "-config", "/app/config.yaml"]
