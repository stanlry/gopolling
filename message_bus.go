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

func newGoroutineSubscription(ch chan Message) *goroutineSubscription {
	return &goroutineSubscription{
		ch: ch,
	}
}

type goroutineSubscription struct {
	ch chan Message
	m  sync.RWMutex
}

func (g *goroutineSubscription) send(message Message) (result bool) {
	g.m.RLock()
	defer g.m.RUnlock()
	defer func() {
		if r := recover(); r != nil {
			result = false
		}
	}()

	g.ch <- message
	return true
}

func (g *goroutineSubscription) Receive() <-chan Message {
	return g.ch
}

func (g *goroutineSubscription) Unsubscribe() error {
	g.m.Lock()
	close(g.ch)
	g.m.Unlock()
	return nil
}

type subQueue struct {
	subs map[int]*goroutineSubscription
	n    int
	m    sync.RWMutex
}

func (q *subQueue) Add(sub *goroutineSubscription) {
	q.m.Lock()
	q.subs[q.n] = sub
	q.n++
	q.m.Unlock()
}

func (q *subQueue) Del(ids []int) {
	q.m.Lock()
	for _, id := range ids {
		delete(q.subs, id)
	}
	q.m.Unlock()
}

type iterFunc func(int, *goroutineSubscription)

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
		var rmi []int
		subq.IterCb(func(i int, sub *goroutineSubscription) {
			if !sub.send(msg) {
				rmi = append(rmi, i)
			}
		})
		subq.Del(rmi)
	} else {
		return ErrNotSubscriber
	}

	return nil
}

func (g *GoroutineBus) Subscribe(roomID string) (Subscription, error) {
	ch := make(chan Message)
	sub := newGoroutineSubscription(ch)
	if val, ok := g.subscribers.Get(roomID); ok {
		subq := val.(*subQueue)
		subq.Add(sub)
	} else {
		subq := subQueue{
			subs: make(map[int]*goroutineSubscription),
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
