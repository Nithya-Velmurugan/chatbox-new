package main

import (
	"chatbox/model"
	"chatbox/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	cs := service.NewChatService()

	r.POST("/join", func(c *gin.Context) {
		var req model.JoinRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		res, err := cs.Join(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/send", func(c *gin.Context) {
		var req model.SendMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		res, err := cs.SendMessage(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/leave", func(c *gin.Context) {
		var req model.LeaveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		res, err := cs.Leave(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/receive/:id", func(c *gin.Context) {
		id := c.Param("id")
		req := model.MessageRequest{ID: id}
		res, err := cs.GetMessage(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusRequestTimeout, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.Run(":8080") // start server on port 8080
}
