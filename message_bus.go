//go:generate mockgen -destination=message_bus_mock_test.go -package=gopolling -source=message_bus.go
package gopolling

import (
	"errors"
)

var (
	ErrNotSubscriber = errors.New("no subscriber")
)

type S map[string]string

type Message struct {
	Data     interface{}
	Error    error
	Selector S
}

type Event struct {
	Data     interface{}
	Selector S
}

type Subscription interface {
	Receive() <-chan Message
	Unsubscribe() error
}

type PubSub interface {
	Publish(string, Message) error
	Subscribe(string) (Subscription, error)
}

type EventQueue interface {
	Enqueue(string, Event)
	// A blocking method wait til receive task
	Dequeue(string) <-chan Event
}

type Loggable interface {
	SetLog(Log)
}

type MessageBus interface {
	PubSub
	EventQueue
	Loggable
}
