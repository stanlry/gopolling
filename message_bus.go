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

type Payload interface {
	Data() interface{}
	Decode(interface{}) error
}

type Msg struct {
	Payload  Payload
	Error    error
	Selector S
}

type Subscription interface {
	Receive() <-chan Msg
}

type PubSub interface {
	Publish(string, interface{}, error, S) error
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

func newGoroutinePayload(data interface{}) Payload {
	return &goroutinePayload{data: data}
}

type goroutinePayload struct {
	data interface{}
}

func (g *goroutinePayload) Data() interface{} {
	return g.data
}

func (g *goroutinePayload) Decode(val interface{}) error {
	val = g.data
	return nil
}

func newGoroutineBus() MessageBus {
	return &GoroutineBus{
		subscribers: cmap.New(),
		queue:       cmap.New(),
	}
}

func newGoroutineSubscription(room, id string, ch chan Msg) *goroutineSubscription {
	return &goroutineSubscription{
		room: room,
		id:   id,
		ch:   ch,
	}
}

type goroutineSubscription struct {
	ch   chan Msg
	room string
	id   string
}

func (g *goroutineSubscription) Receive() <-chan Msg {
	return g.ch
}

type subQueue struct {
	subs map[string]chan Msg
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

type iterFunc func(string, chan Msg)

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

func (g *GoroutineBus) Publish(roomID string, data interface{}, err error, selector S) error {
	if val, ok := g.subscribers.Get(roomID); ok {
		subq := val.(*subQueue)
		msg := Msg{
			Payload:  newGoroutinePayload(data),
			Error:    err,
			Selector: selector,
		}
		subq.IterCb(func(id string, ch chan Msg) {
			ch <- msg
		})
	} else {
		return ErrNotSubscriber
	}

	return nil
}

func (g *GoroutineBus) Subscribe(roomID string) (Subscription, error) {
	ch := make(chan Msg)
	id := xid.New()
	sub := newGoroutineSubscription(roomID, id.String(), ch)
	if val, ok := g.subscribers.Get(roomID); ok {
		subq := val.(*subQueue)
		subq.Add(sub)
	} else {
		subq := subQueue{
			subs: make(map[string]chan Msg),
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
