//go:generate mockgen -destination=message_bus_mock_test.go -package=gopolling -source=message_bus.go
package gopolling

import (
	"errors"
	"github.com/orcaman/concurrent-map"
	"sync"
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

func newGoroutineBus() MessageBus {
	return &GoroutineBus{
		subscribers: cmap.New(),
		queue:       cmap.New(),
	}
}

func newGoroutineSubscription(ch *chan Message) Subscription {
	return &goroutineSubscription{
		ch: ch,
	}
}

type goroutineSubscription struct {
	ch *chan Message
}

func (g *goroutineSubscription) Receive() <-chan Message {
	return *g.ch
}

func (g *goroutineSubscription) Unsubscribe() error {
	close(*g.ch)
	return nil
}

type subQueue struct {
	channels []chan Message
	m        sync.RWMutex
}

func (q *subQueue) Get(i int) chan Message {
	q.m.RLock()
	defer q.m.RUnlock()
	return q.channels[i]
}

func (q *subQueue) Add(ch chan Message) {
	q.m.Lock()
	q.channels = append(q.channels, ch)
	q.m.Unlock()
}

func (q *subQueue) Del(i int) {
	q.m.Lock()
	q.channels = append(q.channels[:i], q.channels[:i+1]...)
	q.m.Unlock()
}

func (q *subQueue) Length() int {
	return len(q.channels)
}

type GoroutineBus struct {
	subscribers cmap.ConcurrentMap
	queue       cmap.ConcurrentMap

	log Log
}

func (g *GoroutineBus) SetLog(l Log) {
	g.log = l
}

func safePushToChannel(ch chan Message, message Message) (result bool) {
	defer func() {
		if r := recover(); r != nil {
			result = false
		}
	}()
	ch <- message
	return true
}

func (g *GoroutineBus) Publish(roomID string, msg Message) error {
	if val, ok := g.subscribers.Get(roomID); ok {
		subq := val.(*subQueue)
		for i := 0; i < subq.Length(); i++ {
			ch := subq.Get(i)
			if !safePushToChannel(ch, msg) {
				subq.Del(i)
			}
		}
	} else {
		return ErrNotSubscriber
	}

	return nil
}

func (g *GoroutineBus) Subscribe(roomID string) (Subscription, error) {
	ch := make(chan Message)
	if val, ok := g.subscribers.Get(roomID); ok {
		subq := val.(*subQueue)
		subq.Add(ch)
	} else {
		subq := subQueue{}
		subq.Add(ch)
		g.subscribers.Set(roomID, &subq)
	}

	return newGoroutineSubscription(&ch), nil
}

func (g *GoroutineBus) Enqueue(roomID string, t Event) {
	if val, ok := g.queue.Get(roomID); ok {
		ch := val.(chan Event)
		ch <- t
	}
}

func (g *GoroutineBus) Dequeue(roomID string) <-chan Event {
	ch := make(chan Event)
	g.queue.Set(roomID, ch)
	return ch
}
