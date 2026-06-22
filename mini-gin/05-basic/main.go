package main

import (
	"fmt"
	"net/http"
)

func main() {
	r := New()

	// 添加全局中间件
	r.Use(func(c *Context) {
		fmt.Println("[Log] Request:", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// 注册根路由
	r.GET("/", func(c *Context) {
		c.JSON(http.StatusOK, map[string]string{
			"message": "Welcome to Mini-Gin!",
		})
	})

	// 注册带参数的路由
	r.GET("/users/:id", func(c *Context) {
		userID := c.Param("id")
		c.JSON(http.StatusOK, map[string]string{
			"user_id": userID,
		})
	})

	fmt.Println("Mini-Gin (Chapter 5) starting on :8080")
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /")
	fmt.Println("  GET  /users/:id")
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
