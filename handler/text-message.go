package handler

import (
	"context"
	"go-meow-test/whatsapp"
	"time"

	"github.com/gin-gonic/gin"
)

func NewTextMessageHandler(r *gin.Engine, wa_client whatsapp.WhatsAppClient, timeout time.Duration) {
	r.POST("/new/message", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		var req Message
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		wa_client.SendMessage(ctx, req.To, req.Body)
		c.JSON(200, gin.H{"status": "ok"})
	})
}
