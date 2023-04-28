package whatsapp

import (
	"context"
	"fmt"
	"os"

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
	sql, err := sqlstore.New("sqlite3", "file:whatsmeow.db?_foreign_keys=on", logger)
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
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			fmt.Println("Message event:", v)
		}
	})
	return &WhatsmeowClient{
		client,
	}, nil
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
	targetJID := types.JID{
		User:   to,
		Server: "s.whatsapp.net",
	}

	_, err := w.client.SendMessage(ctx, targetJID, &proto.Message{
		Conversation: &message,
	})

	return err
}

func (w *WhatsmeowClient) SendGroupMessage(ctx context.Context, group, message string) error {
	targetJID := types.JID{
		User:   group,
		Server: "g.us",
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
	case PresenceComposing:
		p = types.Presence(types.ChatPresenceComposing)
	case PresencePaused:
		p = types.Presence(types.ChatPresencePaused)
	}

	return w.client.SendPresence(p)
}
