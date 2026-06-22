FROM golang:1.23-alpine AS builder

WORKDIR /app

# 使用国内 Go 代理加速模块下载
ENV GOPROXY=https://goproxy.cn,direct

# 复制 go mod 和 sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service .

# 最终镜像
FROM alpine:latest

# 使用国内 Alpine 镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk --no-cache add ca-certificates

WORKDIR /root/

# 从 builder 镜像复制编译好的二进制文件
COPY --from=builder /app/user-service .

EXPOSE 8081

CMD ["./user-service"]
