package whatsapp

import "context"

type Presence string

const (
	PresenceAvailable   Presence = "available"
	PresenceUnavailable Presence = "unavailable"
	PresenceComposing   Presence = "composing"
	PresencePaused      Presence = "paused"
)

type WhatsAppClient interface {
	Connect() error
	Disconnect()

	SendMessage(ctx context.Context, to, message string) error
	SendGroupMessage(ctx context.Context, group, message string) error
	SendPresence(presence Presence) error

	// HandleMessage(handler func())
}
