package gopolling

import (
	"sync"
)

func NewCallback(rid string) Callback {
	return Callback{
		roomID:   rid,
		notified: false,
	}
}

type Callback struct {
	roomID string

	notified bool
	data     interface{}
	err      error
}

func (r *Callback) Notify(data interface{}, err error) {
	r.notified = true
	r.data = data
	r.err = err
}

func (r *Callback) getReplyMsg() *Message {
	return &Message{
		Data:  r.data,
		Error: r.err,
	}
}

type ListenerFunc func(Event, *Callback)

func NewListenerManager(adapter MessageAdapter) ListenerManager {
	return ListenerManager{
		listeners: make(map[string]elm),
		adapter:   adapter,
		log:       &NoOpLog{},
	}
}

type elm struct{}

type ListenerManager struct {
	listeners map[string]elm
	m         sync.Mutex

	log     Log
	adapter MessageAdapter
}

func (m *ListenerManager) listen(roomID string, lf ListenerFunc) {
	for {
		select {
		case task := <-m.adapter.Dequeue(roomID):
			r := NewCallback(roomID)
			lf(task, &r)
			if r.notified {
				msg := r.getReplyMsg()
				msg.Selector = task.Selector
				if err := m.adapter.Publish(roomID, *msg); err != nil {
					m.log.Errorf("fail to publish message, roomID: %v", roomID)
				}
			}
		}
	}
}

func (m *ListenerManager) Subscribe(roomID string, lf ListenerFunc) {
	m.m.Lock()
	if _, ok := m.listeners[roomID]; ok {
		return
	}
	m.listeners[roomID] = elm{}
	m.m.Unlock()

	go m.listen(roomID, lf)
}
