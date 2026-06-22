package main

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    r.GET("/users/:id", func(c *gin.Context) {
        id := c.Param("id")
        c.JSON(http.StatusOK, gin.H{
            "id":   id,
            "name": "Alice",
        })
    })

    r.POST("/users", func(c *gin.Context) {
        c.JSON(http.StatusCreated, gin.H{
            "message": "User created",
        })
    })

    log.Println("User Service starting on :8081")
    r.Run(":8081")
}
