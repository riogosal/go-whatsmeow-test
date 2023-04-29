package main

import (
	"fmt"
	"go-meow-test/chatgpt"
	"go-meow-test/handler"
	"go-meow-test/whatsapp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func bootstrap(wa_chan chan<- whatsapp.WhatsAppClient, gpt_chan chan<- chatgpt.ChatGPTClient, gin_chan chan<- *gin.Engine, error_chan chan<- error) {
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
		wa_chan <- client
	}()
	go func() {
		client := chatgpt.NewOfficialChatGPTClient(10 * time.Second)
		client.WithSystemPrompt("You are a helpful seafood/marine industry expert. Answer in Bahasa Indonesia")
		gpt_chan <- client
	}()

	go func() {
		r := gin.Default()
		gin_chan <- r
	}()
}

func main() {
	godotenv.Load(".env")

	wa_chan := make(chan whatsapp.WhatsAppClient, 1)
	gpt_chan := make(chan chatgpt.ChatGPTClient, 1)
	gin_chan := make(chan *gin.Engine, 1)
	error_chan := make(chan error, 1)

	bootstrap(wa_chan, gpt_chan, gin_chan, error_chan)

	var wa_client whatsapp.WhatsAppClient
	var gpt_client chatgpt.ChatGPTClient
	var r *gin.Engine
	for i := 0; i < 3; i++ {
		select {
		case err := <-error_chan:
			panic(err)
		case client := <-wa_chan:
			wa_client = client
		case client := <-gpt_chan:
			gpt_client = client
		case gin := <-gin_chan:
			r = gin
		}
	}
	defer wa_client.Disconnect()

	fmt.Printf("Bootstrap complete, gin %v, gpt %v, wa %v\n", r, gpt_client, wa_client)

	handler.NewTextMessageHandler(r, wa_client, 10*time.Second)

	r.Run(fmt.Sprintf(":%s", "8080"))
}
