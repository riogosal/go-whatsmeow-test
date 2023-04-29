package whatsapp

type EventName string

const (
	OnMessage EventName = "onmessage"
	OnCall    EventName = "oncall"
)

type EventMapper map[EventName]func()
