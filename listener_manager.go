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

func (r *Callback) Reply(data interface{}, err error) {
	r.notified = true
	r.data = data
	r.err = err
}

type ListenerFunc func(Event, *Callback)

func NewListenerManager(adapter MessageBus, queuePrefix, pubsubPrefix string) ListenerManager {
	return ListenerManager{
		listeners:    make(map[string]elm),
		bus:          adapter,
		log:          &NoOpLog{},
		queuePrefix:  queuePrefix,
		pubsubPrefix: pubsubPrefix,
	}
}

type elm struct{}

type ListenerManager struct {
	listeners map[string]elm
	m         sync.Mutex

	log          Log
	bus          MessageBus
	pubsubPrefix string
	queuePrefix  string
}

func (m *ListenerManager) execListener(roomID string, lf ListenerFunc, ev Event) {
	r := NewCallback(roomID)
	lf(ev, &r)
	if r.notified {
		if err := m.bus.Publish(m.pubsubPrefix+roomID, Message{r.data, r.err, ev.Selector}); err != nil {
			m.log.Errorf("fail to publish message, roomID: %v", roomID)
		}
	}
}

func (m *ListenerManager) listen(roomID string, lf ListenerFunc) {
	for event := range m.bus.Dequeue(m.queuePrefix + roomID) {
		go m.execListener(roomID, lf, event)
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
