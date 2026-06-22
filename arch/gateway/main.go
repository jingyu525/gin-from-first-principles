package main

import (
    "context"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "time"

    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.New()

    // ==================== 可观测性中间件 ====================
    
    // 1. 日志中间件
    r.Use(LoggerMiddleware())
    
    // 2. 链路追踪中间件
    r.Use(TracingMiddleware())
    
    // 3. Prometheus 指标中间件
    r.Use(PrometheusMiddleware())
    
    // 4. 限流中间件（每秒 10 个请求，桶容量 20）
    r.Use(RateLimitMiddleware(10, 20))
    
    // 5. 熔断中间件
    r.Use(CircuitBreakerMiddleware())

    // ==================== 业务中间件 ====================
    
    // JWT 鉴权
    r.Use(JWTAuth())

    // 超时控制
    r.Use(Timeout(5 * time.Second))

    // ==================== 路由定义 ====================
    
    // Prometheus 指标端点
    r.GET("/metrics", MetricsEndpoint())

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

    // 健康检查（不需要鉴权）
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    log.Println("API Gateway starting on :8080")
    log.Println("Endpoints:")
    log.Println("  GET  /health - Health check")
    log.Println("  GET  /metrics - Prometheus metrics")
    log.Println("  ANY  /users/*path - User service proxy")
    log.Println("  ANY  /orders/*path - Order service proxy")
    r.Run(":8080")
}

// createProxy 创建反向代理
func createProxy(target string) *httputil.ReverseProxy {
    url, _ := url.Parse(target)
    return httputil.NewSingleHostReverseProxy(url)
}

// JWTAuth JWT 鉴权中间件
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 健康检查接口不需要鉴权
        if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
            c.Next()
            return
        }

        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }

        // 简化示例：实际应该验证 JWT token
        if len(token) < 7 || token[:7] != "Bearer " {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        c.Next()
    }
}

// Timeout 超时控制中间件
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
            c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{"error": "Gateway Timeout"})
        }
    }
}
