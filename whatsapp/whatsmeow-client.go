package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"go-meow-test/chatgpt"
	"os"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type WhatsmeowClient struct {
	client *whatsmeow.Client
}

func NewWhatsMeowClient() (WhatsAppClient, error) {
	logger := waLog.Stdout("Database", "DEBUG", true)
	sql, err := sqlstore.New("sqlite3", fmt.Sprintf("file:%s.db?_foreign_keys=on", os.Getenv("WHATSAPP_DB_NAME")), logger)
	if err != nil {
		return nil, err
	}
	container, err := sql.GetFirstDevice()
	if err != nil {
		return nil, err
	}
	client := whatsmeow.NewClient(
		container,
		logger,
	)
	return &WhatsmeowClient{
		client,
	}, nil
}

var bot_prompt = regexp.MustCompile(`^\!bot`)

func (w *WhatsmeowClient) SetEventsHandler(chatgpt chatgpt.ChatGPTClient) {
	w.client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.IsFromMe {
				return
			}
			if v.Message.Conversation == nil {
				return
			}
			if !bot_prompt.Match([]byte(*v.Message.Conversation)) {
				return
			}
			if target := os.Getenv("WHATSAPP_TARGET_JID"); v.Info.Sender.String() == target {
				w.client.SendChatPresence(v.Info.Sender, types.ChatPresenceComposing, types.ChatPresenceMediaText)
				// only run chatgpt if the message is from the target
				if result, err := chatgpt.ChatCompletion(*v.Message.Conversation); err != nil {
					w.SendMessage(context.Background(), target, "Bot chat is unavailable right now. Sorry!")
				} else {
					w.SendMessage(context.Background(), target, result)
				}
				w.client.SendChatPresence(v.Info.Sender, types.ChatPresencePaused, types.ChatPresenceMediaText)
				w.SendPresence(PresenceUnavailable)

			}
		}
	})
}

func (w *WhatsmeowClient) Disconnect() {
	w.client.Disconnect()
}

func (w *WhatsmeowClient) Connect() error {
	if w.client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := w.client.GetQRChannel(context.Background())
		err := w.client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err := w.client.Connect()
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func (w *WhatsmeowClient) SendMessage(ctx context.Context, to, message string) error {
	split_str := strings.Split(to, "@")
	if len(split_str) != 2 {
		return errors.New("invalid jid")
	}
	targetJID := types.JID{
		User:   split_str[0],
		Server: split_str[1],
	}

	_, err := w.client.SendMessage(ctx, targetJID, &proto.Message{
		Conversation: &message,
	})

	return err
}

func (w *WhatsmeowClient) SendPresence(presence Presence) error {
	var p types.Presence
	switch presence {
	case PresenceAvailable:
		p = types.PresenceAvailable
	case PresenceUnavailable:
		p = types.PresenceUnavailable
	}

	return w.client.SendPresence(p)
}
