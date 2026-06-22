# 第十三章：避坑指南（分布式下的 Gin）

## 1. Context 的生命周期陷阱

### 1.1 为什么不能在 Goroutine 里直接使用 Context？
**场景**：在 Handler 中启动一个 Goroutine。
```go
func Handler(c *gin.Context) {
    go func() {
        // ❌ 危险：c 可能已经被回收或复用
        userID := c.GetString("user_id")
        fmt.Println(userID)
    }()
    c.JSON(200, gin.H{"status": "ok"})
}
```

**原因**：
- `c` 是从 `sync.Pool` 中复用的。
- 当 Handler 返回后，`c` 会被重置并归还到池中。
- Goroutine 中访问的 `c` 可能已经被其他请求复用。

### 1.2 正确的做法
```go
func Handler(c *gin.Context) {
    // ✅ 正确：先复制需要的值
    userID := c.GetString("user_id")
    
    go func(uid string) {
        fmt.Println(uid)  // 使用局部变量
    }(userID)
    
    c.JSON(200, gin.H{"status": "ok"})
}
```

---

## 2. TraceID 的传播

### 2.1 为什么需要 TraceID？
**场景**：一个请求经过多个服务，如何追踪完整链路？
```
Client → Gateway → User Service → Order Service → Payment Service
```
**需求**：在日志中打印同一个 TraceID。

### 2.2 如何在 Gateway 和 Service 间传递追踪信息
#### 方案 A：HTTP Header
```go
// Gateway 中生成 TraceID
traceID := uuid.New().String()
c.Request.Header.Set("X-Trace-ID", traceID)

// Service 中读取 TraceID
traceID := c.GetHeader("X-Trace-ID")
log.Printf("[%s] Handling request", traceID)
```

#### 方案 B：OpenTelemetry（推荐）
使用标准库自动传播 TraceID。

---

## 3. 超时传递的一致性

### 3.1 网关超时 vs 服务超时的推导
**场景**：
- Gateway 设置超时 5 秒。
- Service A 调用 Service B，设置超时 10 秒。

**问题**：Service A 还在等待，但 Gateway 已经返回超时。

### 3.2 正确的做法：传递 context.Context
```go
// Gateway 中设置超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 将 ctx 传递给下游服务
req := http.Request{
    URL:    &url.URL{Path: "/api/orders"},
    Header: make(http.Header),
}
req = req.WithContext(ctx)

resp, err := http.DefaultClient.Do(req)
```

### 3.3 在 Gin 中使用 context.Context
```go
func Handler(c *gin.Context) {
    // 获取请求的 context
    ctx := c.Request.Context()
    
    // 调用下游服务时传递 ctx
    req, _ := http.NewRequestWithContext(ctx, "GET", "http://order-service/orders", nil)
    resp, err := http.DefaultClient.Do(req)
    
    // 如果 Gateway 超时，ctx 会被取消，req 也会自动取消
}
```

---

## 4. 其他常见坑

### 4.1 在中间件中使用 c.JSON() 后忘记 return
```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.GetHeader("Token") == "" {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return  // ✅ 必须 return，否则会继续执行后续中间件
        }
        c.Next()
    }
}
```

### 4.2 在 Goroutine 中使用 c.JSON()
```go
// ❌ 错误：Goroutine 中不能直接写响应
func Handler(c *gin.Context) {
    go func() {
        time.Sleep(5 * time.Second)
        c.JSON(200, gin.H{"status": "done"})  // 主函数可能已经返回
    }()
    c.JSON(202, gin.H{"status": "accepted"})
}

// ✅ 正确：使用 channel 通信
func Handler(c *gin.Context) {
    done := make(chan struct{})
    go func() {
        time.Sleep(5 * time.Second)
        close(done)
    }()
    
    <-done
    c.JSON(200, gin.H{"status": "done"})
}
```

---

## 5. 总结：你的位置在哪里

### 5.1 回顾全链路
```
字节流 (Bytes)
    ↓
TCP 连接 (net.Listen)
    ↓
HTTP 协议解析 (net/http)
    ↓
路由抽象 (Map -> Radix Tree)
    ↓
上下文管理 (Context & sync.Pool)
    ↓
中间件责任链 (Chain & Abort)
    ↓
微服务拆分 (Service Layer)
    ↓
流量入口 (API Gateway)
    ↓
服务治理 (Service Mesh)
```

### 5.2 你现在能做什么？
- ✅ 手写 Mini-Gin 框架
- ✅ 理解 Gin 的核心原理
- ✅ 设计微服务架构
- ✅ 实现 API Gateway
- ✅ 避免分布式下的常见坑

**恭喜你，已经完成从第一性原理到分布式架构的全链路学习！** 🎉

---

## 附录：完整代码索引

- **Mini-Gin 源码**：`mini-gin/`
- **API Gateway 示例**：`arch/gateway/`
- **用户服务示例**：`arch/user-service/`
- **订单服务示例**：`arch/order-service/`

**Happy Coding!** 🚀
