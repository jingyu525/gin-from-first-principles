package main

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    r.GET("/orders/:id", func(c *gin.Context) {
        id := c.Param("id")
        c.JSON(http.StatusOK, gin.H{
            "id":     id,
            "item":   "Book",
            "amount": 29.99,
        })
    })

    r.POST("/orders", func(c *gin.Context) {
        c.JSON(http.StatusCreated, gin.H{
            "message": "Order created",
        })
    })

    log.Println("Order Service starting on :8082")
    r.Run(":8082")
}
