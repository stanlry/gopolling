package adapter

import (
	cmap "github.com/orcaman/concurrent-map"
	"github.com/stanlry/gopolling"
	"sync"
)

func NewGoroutineAdapter() gopolling.MessageBus {
	return &GoroutineAdapter{
		subscribers: cmap.New(),
		queue:       cmap.New(),
	}
}

func newGoroutineSubscription(ch *chan gopolling.Message) gopolling.Subscription {
	return &goroutineSubscription{
		ch: ch,
	}
}

type goroutineSubscription struct {
	ch *chan gopolling.Message
}

func (g *goroutineSubscription) Receive() <-chan gopolling.Message {
	return *g.ch
}

func (g *goroutineSubscription) Unsubscribe() error {
	close(*g.ch)
	return nil
}

type subQueue struct {
	channels []chan gopolling.Message
	m        sync.RWMutex
}

func (q *subQueue) Get(i int) chan gopolling.Message {
	q.m.RLock()
	defer q.m.RUnlock()
	return q.channels[i]
}

func (q *subQueue) Add(ch chan gopolling.Message) {
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

type GoroutineAdapter struct {
	subscribers cmap.ConcurrentMap
	queue       cmap.ConcurrentMap

	log gopolling.Log
}

func (g *GoroutineAdapter) SetLog(l gopolling.Log) {
	g.log = l
}

func safePushToChannel(ch chan gopolling.Message, message gopolling.Message) (result bool) {
	defer func() {
		if r := recover(); r != nil {
			result = false
		}
	}()
	ch <- message
	return true
}

func (g *GoroutineAdapter) Publish(roomID string, msg gopolling.Message) error {
	if val, ok := g.subscribers.Get(roomID); ok {
		subq := val.(*subQueue)
		for i := 0; i < subq.Length(); i++ {
			ch := subq.Get(i)
			if !safePushToChannel(ch, msg) {
				subq.Del(i)
			}
		}
	} else {
		return gopolling.ErrNotSubscriber
	}

	return nil
}

func (g *GoroutineAdapter) Subscribe(roomID string) (gopolling.Subscription, error) {
	ch := make(chan gopolling.Message)
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

func (g *GoroutineAdapter) Enqueue(roomID string, t gopolling.Event) {
	if val, ok := g.queue.Get(roomID); ok {
		ch := val.(chan gopolling.Event)
		ch <- t
	}
}

func (g *GoroutineAdapter) Dequeue(roomID string) <-chan gopolling.Event {
	ch := make(chan gopolling.Event)
	g.queue.Set(roomID, ch)
	return ch
}
