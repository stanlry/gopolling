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

func NewPollingManager(adapter MessageAdapter) PollingManager {
	return PollingManager{
		adapter: adapter,
		timeout: defaultTimeout,
		log:     &NoOpLog{},
	}
}

type PollingManager struct {
	adapter MessageAdapter
	timeout time.Duration

	log Log
}

func (m *PollingManager) WaitForNotice(ctx context.Context, roomID string, data interface{}, sel S) (interface{}, error) {
	sub, err := m.adapter.Subscribe(roomID)
	if err != nil {
		return nil, err
	}

	tick := time.Tick(m.timeout)

	// enqueue task
	tk := Event{
		Data:     data,
		Selector: sel,
	}
	m.adapter.Enqueue(roomID, tk)

wait:
	select {
	case <-ctx.Done():
		if err := sub.Unsubscribe(); err != nil {
			m.log.Errorf("fail to unsubscribe from adapter, roomID: %v", roomID)
		}
		return nil, ErrCancelled
	case <-tick:
		if err := sub.Unsubscribe(); err != nil {
			m.log.Errorf("fail to unsubscribe from adapter, roomID: %v", roomID)
		}
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
