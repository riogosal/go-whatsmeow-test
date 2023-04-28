package main

import (
	"context"
	"fmt"
	"go-meow-test/chatgpt"
	"go-meow-test/domain"
	"go-meow-test/whatsapp"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func bootstrap() (wa_client whatsapp.WhatsAppClient, chatgpt_client chatgpt.ChatGPTClient) {
	error_chan := make(chan error)
	wa_ready := make(chan struct{})
	chatgpt_ready := make(chan struct{})

	var wa whatsapp.WhatsAppClient
	var gpt chatgpt.ChatGPTClient

	go func() {
		client, err := whatsapp.NewWhatsMeowClient()
		if err != nil {
			error_chan <- err
			return
		}
		if err := client.Connect(); err != nil {
			error_chan <- err
			return
		}
		client.SendPresence(whatsapp.PresenceUnavailable)
		wa = client

		wa_ready <- struct{}{}
	}()
	go func() {
		client := chatgpt.NewOfficialChatGPTClient(10 * time.Second)
		client.WithSystemPrompt("You are a helpful seafood/marine industry expert. Answer in Bahasa Indonesia")
		gpt = client

		chatgpt_ready <- struct{}{}
	}()

	for i := 0; i < 2; i++ {
		select {
		case err := <-error_chan:
			panic(err)
		case <-wa_ready:
			fmt.Println("whatsapp client ready")
		case <-chatgpt_ready:
			fmt.Println("chatgpt client ready")
		}
	}
	return wa, gpt
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	godotenv.Load(".env")
	wa_client, _ := bootstrap()

	r := gin.Default()
	r.POST("/new/message", func(c *gin.Context) {
		var req domain.Message
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		wa_client.SendMessage(context.Background(), req.To, req.Body)
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/new/group/message", func(c *gin.Context) {
		var req domain.Message
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		wa_client.SendGroupMessage(context.Background(), req.To, req.Body)
		c.JSON(200, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:    ":8000",
		Handler: r,
	}

	go func() {
		if err := r.Run(":8000"); err != nil {
			panic(err)
		}
	}()

	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	fmt.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
	wa_client.Disconnect()

	fmt.Println("Server and whatsapp client exiting")
}
