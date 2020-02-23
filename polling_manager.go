package gopolling

import (
	"context"
	"errors"
	"github.com/rs/xid"
	"time"
)

var (
	ErrTimeout     = errors.New("request timeout")
	ErrCancelled   = errors.New("request cancelled")
	defaultTimeout = 120 * time.Second
)

func NewPollingManager(adapter MessageBus, t time.Duration, queuePrefix, pubsubPrefix string) PollingManager {
	timeout := defaultTimeout
	if t != 0 {
		timeout = t
	}
	return PollingManager{
		bus:          adapter,
		timeout:      timeout,
		log:          &NoOpLog{},
		queuePrefix:  queuePrefix,
		pubsubPrefix: pubsubPrefix,
	}
}

type PollingManager struct {
	bus     MessageBus
	timeout time.Duration

	log          Log
	queuePrefix  string
	pubsubPrefix string
}

const idKey = "_gopolling_id"

func (m *PollingManager) WaitForNotice(ctx context.Context, channel string, data interface{}, sel S) (interface{}, error) {
	rChannel := m.pubsubPrefix + channel
	sub, err := m.bus.Subscribe(rChannel)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := m.bus.Unsubscribe(sub); err != nil {
			m.log.Errorf("unsubscribe unsuccessful, error: ", err)
		}
	}()

	tick := time.NewTicker(m.timeout)
	defer tick.Stop()

	// generate event ID
	id := xid.New().String()
	sel[idKey] = id
	// enqueue event
	qChan := m.queuePrefix + channel
	tk := Event{
		Channel:  qChan,
		Data:     data,
		Selector: sel,
	}
	m.bus.Enqueue(qChan, tk)
	delete(sel, idKey)

wait:
	select {
	case <-ctx.Done():
		return nil, context.Canceled
	case <-tick.C:
		return nil, ErrTimeout
	case msg := <-sub.Receive():
		// if msg is specified with event ID
		if val, ok := msg.Selector[idKey]; ok {
			if val == id {
				return msg.Data, msg.Error
			} else {
				goto wait
			}
		}

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
