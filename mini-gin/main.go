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

    // 注册路由
    r.GET("/", func(c *Context) {
        c.JSON(http.StatusOK, map[string]string{
            "message": "Welcome to Mini-Gin!",
        })
    })

    r.GET("/users/:id", func(c *Context) {
        userID := c.Param("id")
        c.JSON(http.StatusOK, map[string]string{
            "user_id": userID,
        })
    })

    r.POST("/users", func(c *Context) {
        c.JSON(http.StatusCreated, map[string]string{
            "message": "User created",
        })
    })

    // 启动服务器
    fmt.Println("Mini-Gin server starting on :8080")
    if err := r.Run(":8080"); err != nil {
        fmt.Printf("Server error: %v\n", err)
    }
}
