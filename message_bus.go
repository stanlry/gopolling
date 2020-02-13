//go:generate mockgen -destination=message_bus_mock_test.go -package=gopolling -source=message_bus.go
package gopolling

import (
	"errors"
	"github.com/orcaman/concurrent-map"
	"github.com/rs/xid"
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

func newGoroutineSubscription(room, id string, ch chan Message) *goroutineSubscription {
	return &goroutineSubscription{
		room: room,
		id:   id,
		ch:   ch,
	}
}

type goroutineSubscription struct {
	ch   chan Message
	room string
	id   string
}

func (g *goroutineSubscription) Receive() <-chan Message {
	return g.ch
}

type subQueue struct {
	subs map[string]chan Message
	m    sync.RWMutex
}

func (q *subQueue) Add(sub *goroutineSubscription) {
	q.m.Lock()
	q.subs[sub.id] = sub.ch
	q.m.Unlock()
}

func (q *subQueue) Del(id string) {
	q.m.Lock()
	delete(q.subs, id)
	q.m.Unlock()
}

type iterFunc func(string, chan Message)

func (q *subQueue) IterCb(fn iterFunc) {
	q.m.RLock()
	for i, v := range q.subs {
		fn(i, v)
	}
	q.m.RUnlock()
}

type GoroutineBus struct {
	subscribers cmap.ConcurrentMap
	queue       cmap.ConcurrentMap

	log Log
}

func (g *GoroutineBus) Unsubscribe(sub Subscription) error {
	gsub := sub.(*goroutineSubscription)
	if val, ok := g.subscribers.Get(gsub.room); ok {
		subq := val.(*subQueue)
		subq.Del(gsub.id)
	}
	return nil
}

func (g *GoroutineBus) SetLog(l Log) {
	g.log = l
}

func (g *GoroutineBus) Publish(roomID string, msg Message) error {
	if val, ok := g.subscribers.Get(roomID); ok {
		subq := val.(*subQueue)
		subq.IterCb(func(id string, ch chan Message) {
			ch <- msg
		})
	} else {
		return ErrNotSubscriber
	}

	return nil
}

func (g *GoroutineBus) Subscribe(roomID string) (Subscription, error) {
	ch := make(chan Message)
	id := xid.New()
	sub := newGoroutineSubscription(roomID, id.String(), ch)
	if val, ok := g.subscribers.Get(roomID); ok {
		subq := val.(*subQueue)
		subq.Add(sub)
	} else {
		subq := subQueue{
			subs: make(map[string]chan Message),
		}
		subq.Add(sub)
		g.subscribers.Set(roomID, &subq)
	}

	return sub, nil
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
