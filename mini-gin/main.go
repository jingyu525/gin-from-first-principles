package main

import (
    "fmt"
    "net/http"
)

func main() {
    r := New()

    // 添加全局中间件
    r.Use(func(c *Context) {
        fmt.Println("Global middleware: Before request")
        c.Next()
        fmt.Println("Global middleware: After request")
    })

    // 获取默认路由组
    g := r.Default()

    // 注册根路由
    g.GET("/", func(c *Context) {
        c.JSON(http.StatusOK, map[string]string{
            "message": "Welcome to Mini-Gin!",
        })
    })

    // V1 路由组（旧版 API）
    v1 := g.Group("/v1")
    v1.Use(func(c *Context) {
        fmt.Println("V1 middleware")
        c.Next()
    })
    v1.GET("/users/:id", func(c *Context) {
        userID := c.Param("id")
        c.JSON(http.StatusOK, map[string]string{
            "version":  "v1",
            "user_id": userID,
        })
    })

    // V2 路由组（新版 API）
    v2 := g.Group("/v2")
    v2.Use(func(c *Context) {
        fmt.Println("V2 middleware")
        c.Next()
    })
    v2.GET("/users/:id", func(c *Context) {
        userID := c.Param("id")
        c.JSON(http.StatusOK, map[string]interface{}{
            "version":  "v2",
            "user_id": userID,
            "email":   "user@example.com",
        })
    })

    // 启动服务器
    fmt.Println("Mini-Gin server starting on :8080")
    fmt.Println("Endpoints:")
    fmt.Println("  GET  /")
    fmt.Println("  GET  /v1/users/:id")
    fmt.Println("  GET  /v2/users/:id")
    if err := r.Run(":8080"); err != nil {
        fmt.Printf("Server error: %v\n", err)
    }
}
