package gopolling

import (
	"context"
	"errors"
	"time"
)

var (
	ErrTimeout     = errors.New("timeout")
	ErrCancelled   = errors.New("cancelled")
	defaultTimeout = 120 * time.Second
)

func NewPollingManager(adapter MessageBus) PollingManager {
	return PollingManager{
		bus:     adapter,
		timeout: defaultTimeout,
		log:     &NoOpLog{},
	}
}

type PollingManager struct {
	bus     MessageBus
	timeout time.Duration

	log Log
}

func (m *PollingManager) WaitForNotice(ctx context.Context, roomID string, data interface{}, sel S) (interface{}, error) {
	sub, err := m.bus.Subscribe(roomID)
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	tick := time.Tick(m.timeout)

	// enqueue task
	tk := Event{
		Data:     data,
		Selector: sel,
	}
	m.bus.Enqueue(roomID, tk)

wait:
	select {
	case <-ctx.Done():
		return nil, ErrCancelled
	case <-tick:
		return nil, ErrTimeout
	case msg := <-sub.Receive():
		if !compareSelectors(msg.Selector, sel) {
			goto wait
		}
		return msg.Data, msg.Error
	}
}

func compareSelectors(a, b S) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if val, ok := b[k]; ok && val == v {
			continue
		} else {
			return false
		}
	}

	return true
}
