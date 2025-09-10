# 使用官方Go镜像作为构建环境
FROM golang:1.20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 使用轻量级镜像作为运行环境
FROM alpine:latest

# 安装ca证书（HTTPS请求需要）
RUN apk --no-cache add ca-certificates

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 更改文件所有者
RUN chown -R appuser:appgroup /root/

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"]
