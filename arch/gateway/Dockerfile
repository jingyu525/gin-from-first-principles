FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制 go mod 和 sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway .

# 最终镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从 builder 镜像复制编译好的二进制文件
COPY --from=builder /app/gateway .

EXPOSE 8080

CMD ["./gateway"]
