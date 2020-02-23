package gopolling

//go:generate mockgen -destination=message_bus_mock_test.go -package=gopolling -source=message_bus.go

import (
	"errors"
	"github.com/orcaman/concurrent-map"
	"github.com/rs/xid"
)

var (
	ErrNotSubscriber = errors.New("no subscriber")
)

// A shortcut to define string map
type S map[string]string

// Message struct defines the payload which being sent through message bus from notifier to the waiting client
type Message struct {
	Channel  string
	Data     interface{}
	Error    error
	Selector S
}

// Event struct is the payload being sent from client to the listening subscriber
type Event struct {
	Channel  string
	Data     interface{}
	Selector S
}

type Subscription interface {
	Receive() <-chan Message
}

func NewDefaultSubscription(channel, id string, ch chan Message) *DefaultSubscription {
	return &DefaultSubscription{
		Channel: channel,
		ID:      id,
		ch:      ch,
	}
}

type DefaultSubscription struct {
	ch      chan Message
	Channel string
	ID      string
}

func (g *DefaultSubscription) Receive() <-chan Message {
	return g.ch
}

type PubSub interface {
	Publish(string, Message) error
	Subscribe(string) (Subscription, error)
	Unsubscribe(Subscription) error
}

type EventQueue interface {
	Enqueue(string, Event)
	// A blocking method wait til receive task
	Dequeue(string) <-chan Event
}

type MessageBus interface {
	PubSub
	EventQueue
	Loggable
}

func newGoroutineBus() MessageBus {
	return &GoroutineBus{
		subscribers: cmap.New(),
		queue:       cmap.New(),
	}
}

type GoroutineBus struct {
	subscribers cmap.ConcurrentMap
	queue       cmap.ConcurrentMap

	log Log
}

func (g *GoroutineBus) Unsubscribe(sub Subscription) error {
	gsub := sub.(*DefaultSubscription)
	if val, ok := g.subscribers.Get(gsub.Channel); ok {
		subq := val.(*cmap.ConcurrentMap)
		subq.Remove(gsub.ID)
	}
	return nil
}

func (g *GoroutineBus) SetLog(l Log) {
	g.log = l
}

func (g *GoroutineBus) Publish(channel string, msg Message) error {
	if val, ok := g.subscribers.Get(channel); ok {
		subq := val.(*cmap.ConcurrentMap)
		subq.IterCb(func(id string, v interface{}) {
			ch := v.(chan Message)
			ch <- msg
		})
	} else {
		return ErrNotSubscriber
	}

	return nil
}

func (g *GoroutineBus) Subscribe(channel string) (Subscription, error) {
	ch := make(chan Message)
	id := xid.New().String()
	sub := NewDefaultSubscription(channel, id, ch)
	if val, ok := g.subscribers.Get(channel); ok {
		subMap := val.(*cmap.ConcurrentMap)
		subMap.Set(id, ch)
	} else {
		subMap := cmap.New()
		subMap.Set(id, ch)
		g.subscribers.Set(channel, &subMap)
	}

	return sub, nil
}

func (g *GoroutineBus) Enqueue(channel string, t Event) {
	if val, ok := g.queue.Get(channel); ok {
		ch := val.(chan Event)
		ch <- t
	}
}

func (g *GoroutineBus) Dequeue(channel string) <-chan Event {
	ch := make(chan Event)
	g.queue.Set(channel, ch)
	return ch
}
