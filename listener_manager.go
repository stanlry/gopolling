package gopolling

import (
	"sync"
)

func NewCallback() Callback {
	return Callback{
		notified: false,
	}
}

type Callback struct {
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

func (m *ListenerManager) execListener(channel string, lf ListenerFunc, ev Event) {
	r := NewCallback()
	lf(ev, &r)
	if r.notified {
		pChan := m.pubsubPrefix + channel
		if err := m.bus.Publish(pChan, Message{pChan, r.data, r.err, ev.Selector}); err != nil {
			m.log.Errorf("fail to publish message, Channel: %v", channel)
		}
	}
}

func (m *ListenerManager) listen(channel string, lf ListenerFunc) {
	for event := range m.bus.Dequeue(m.queuePrefix + channel) {
		go m.execListener(channel, lf, event)
	}
}

func (m *ListenerManager) Subscribe(channel string, lf ListenerFunc) {
	m.m.Lock()
	if _, ok := m.listeners[channel]; ok {
		return
	}
	m.listeners[channel] = elm{}
	m.m.Unlock()

	go m.listen(channel, lf)
}
