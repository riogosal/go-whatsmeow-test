package main

import (
	"fmt"
	"go-meow-test/chatgpt"
	"go-meow-test/handler"
	"go-meow-test/whatsapp"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

func bootstrap() (whatsapp.WhatsAppClient, chatgpt.ChatGPTClient, *gin.Engine) {
	wa_chan := make(chan whatsapp.WhatsAppClient, 1)
	gpt_chan := make(chan chatgpt.ChatGPTClient, 1)
	gin_chan := make(chan *gin.Engine, 1)

	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		client, err := whatsapp.NewWhatsMeowClient()
		if err != nil {
			return err
		}
		wa_chan <- client
		return nil
	})
	errgrp.Go(func() error {
		client := chatgpt.NewOfficialChatGPTClient(10 * time.Second)
		gpt_chan <- client
		return nil
	})
	errgrp.Go(func() error {
		if os.Getenv("APP_ENV") == "production" {
			gin.SetMode(gin.ReleaseMode)
		}
		r := gin.Default()
		gin_chan <- r
		return nil
	})

	if err := errgrp.Wait(); err != nil {
		panic(err)
	}

	return <-wa_chan, <-gpt_chan, <-gin_chan
}

func main() {
	godotenv.Load(".env")

	wa_client, gpt_client, r := bootstrap()
	gpt_client.WithSystemPrompt("You are a helpful seafood/marine industry expert. Answer in Bahasa Indonesia")

	wa_client.SetEventsHandler(gpt_client)
	if err := wa_client.Connect(); err != nil {
		panic(err)
	}
	defer wa_client.Disconnect()

	handler.NewTextMessageHandler(r, wa_client, 10*time.Second)

	r.Run(fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")))
}
