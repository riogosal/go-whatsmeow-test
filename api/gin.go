package api

import (
	"context"
	"go-meow-test/chatgpt"
	"go-meow-test/whatsapp"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Router    *gin.Engine
	WAClient  whatsapp.WhatsAppClient
	GPTClient chatgpt.ChatGPTClient

	Timeout time.Duration
}

func (h *Handler) HandleMessage(wa_client whatsapp.WhatsAppClient) {
	h.Router.POST("/new/message", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), h.Timeout)
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

func (h *Handler) HandleGroupMessage(wa_client whatsapp.WhatsAppClient) {
	h.Router.POST("/new/group/message", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), h.Timeout)
		defer cancel()

		var req Message
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		wa_client.SendGroupMessage(ctx, req.To, req.Body)
		c.JSON(200, gin.H{"status": "ok"})
	})
}

func NewHandler(timeout time.Duration) *Handler {
	return &Handler{
		Router:  gin.Default(),
		Timeout: timeout,
	}
}
