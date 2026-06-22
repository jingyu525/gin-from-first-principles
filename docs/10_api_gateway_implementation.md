# 第十章：API Gateway（手写实战）

## 1. 网关的第一性原理

### 1.1 定义：系统的唯一入口
**问题**：客户端如何知道有多个服务？
```
❌ 错误：客户端直接调用多个服务
Client → User Service (8081)
Client → Order Service (8082)
Client → Payment Service (8083)

✅ 正确：客户端只访问网关
Client → Gateway (8080) → 转发到各个服务
```

### 1.2 网关的核心职责
1.  **鉴权**：验证 Token，拒绝非法请求
2.  **限流**：防止单个客户端打垮服务
3.  **路由转发**：根据路径转发到对应服务
4.  **超时控制**：防止下游服务慢响应拖垮网关

---

## 2. 用 Gin 实现反向代理

### 2.1 使用 httputil.ReverseProxy
```go
package main

import (
    "net/http"
    "net/http/httputil"
    "net/url"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    // 用户服务代理
    userSvc := "http://localhost:8081"
    proxyUser := createProxy(userSvc)
    r.Any("/users/*path", func(c *gin.Context) {
        proxyUser.ServeHTTP(c.Writer, c.Request)
    })
    
    // 订单服务代理
    orderSvc := "http://localhost:8082"
    proxyOrder := createProxy(orderSvc)
    r.Any("/orders/*path", func(c *gin.Context) {
        proxyOrder.ServeHTTP(c.Writer, c.Request)
    })
    
    r.Run(":8080")
}

func createProxy(target string) *httputil.ReverseProxy {
    url, _ := url.Parse(target)
    return httputil.NewSingleHostReverseProxy(url)
}
```

---

## 3. 实现 JWT 鉴权中间件

### 3.1 中间件代码
```go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
            return
        }
        
        // 验证 JWT（省略具体逻辑）
        if !validateJWT(token) {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
            return
        }
        
        c.Next()
    }
}
```

### 3.2 使用中间件
```go
r := gin.Default()
r.Use(JWTAuth())  // 全局鉴权
```

---

## 4. 实现超时控制中间件

### 4.1 使用 context.WithTimeout
```go
func Timeout(duration time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
        defer cancel()
        c.Request = c.Request.WithContext(ctx)
        
        done := make(chan struct{})
        go func() {
            c.Next()
            done <- struct{}{}
        }()
        
        select {
        case <-done:
            // 正常完成
        case <-ctx.Done():
            c.AbortWithStatusJSON(504, gin.H{"error": "Gateway Timeout"})
        }
    }
}
```

---

## 5. 完整的 Gateway 代码

```go
package main

import (
    "time"
    "net/http"
    "net/http/httputil"
    "net/url"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    // 全局中间件
    r.Use(JWTAuth())
    r.Use(Timeout(5 * time.Second))
    
    // 路由转发
    userSvc := "http://localhost:8081"
    proxyUser := createProxy(userSvc)
    r.Any("/users/*path", func(c *gin.Context) {
        proxyUser.ServeHTTP(c.Writer, c.Request)
    })
    
    orderSvc := "http://localhost:8082"
    proxyOrder := createProxy(orderSvc)
    r.Any("/orders/*path", func(c *gin.Context) {
        proxyOrder.ServeHTTP(c.Writer, c.Request)
    })
    
    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    r.Run(":8080")
}

// createProxy, JWTAuth, Timeout 函数省略...
```

---

## 6. 下一章预告

API Gateway 实现完成！接下来我们将 **祛魅 Service Mesh**：
- 东西流量 vs 南北流量
- Sidecar 模式的推导

**准备进入云原生时代！** ☁️
