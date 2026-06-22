package main

import (
    "log"
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

    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    log.Println("API Gateway starting on :8080")
    r.Run(":8080")
}

// createProxy 创建反向代理
func createProxy(target string) *httputil.ReverseProxy {
    url, _ := url.Parse(target)
    return httputil.NewSingleHostReverseProxy(url)
}
