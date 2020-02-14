package gopolling

import (
	"context"
	"strings"
	"time"
)

const (
	queuePrefix  = "gpqueue:"
	pubsubPrefix = "gppubsub:"
)

type Option struct {
	// Retention period indicate the TTL time in second for the message, default 60s
	Retention int

	// Long polling client retention (default 120s)
	Timeout time.Duration

	// Message Bus
	Bus MessageBus

	// Message Buffer
	Buffer MessageBuffer

	// Logging implementation
	Logger Log

	PubSubPrefix string

	QueuePrefix string
}

var DefaultOption = Option{
	Retention:    600,
	Timeout:      120 * time.Second,
	Bus:          newGoroutineBus(),
	Buffer:       newMemoryBuffer(),
	PubSubPrefix: pubsubPrefix,
	QueuePrefix:  queuePrefix,
	Logger:       &NoOpLog{},
}

func buildOption(option Option) Option {
	op := DefaultOption
	if option.Timeout > 0 {
		op.Timeout = option.Timeout
	}
	if option.Logger != nil {
		op.Logger = option.Logger
	}
	if option.Bus != nil {
		op.Bus = option.Bus
	}
	if option.Buffer != nil {
		op.Buffer = option.Buffer
	}
	if option.Retention != 0 {
		op.Retention = option.Retention
	}
	if len(strings.TrimSpace(option.QueuePrefix)) != 0 {
		op.QueuePrefix = option.QueuePrefix
	}
	if len(strings.TrimSpace(option.PubSubPrefix)) != 0 {
		op.PubSubPrefix = option.PubSubPrefix
	}
	return op
}

func New(opt Option) *GoPolling {
	var gp GoPolling

	option := buildOption(opt)

	gp.pollingMgr = NewPollingManager(option.Bus, option.Timeout, option.QueuePrefix, option.PubSubPrefix)
	gp.listenerMgr = NewListenerManager(option.Bus, option.QueuePrefix, option.PubSubPrefix)

	gp.pollingMgr.log = option.Logger
	gp.listenerMgr.log = option.Logger

	gp.bus = option.Bus
	gp.bus.SetLog(option.Logger)
	gp.buffer = option.Buffer
	gp.retention = option.Retention
	gp.pubsubPrefix = option.PubSubPrefix

	return &gp
}

type GoPolling struct {
	bus          MessageBus
	pollingMgr   PollingManager
	listenerMgr  ListenerManager
	buffer       MessageBuffer
	retention    int
	pubsubPrefix string
}

func (g *GoPolling) WaitForNotice(ctx context.Context, roomID string, data interface{}) (interface{}, error) {
	if g.buffer != nil {
		key := bufferKey{roomID, S{}}
		hashedKey := getKeyHash(key)
		if val, ok := g.buffer.Find(hashedKey); ok {
			return val.Data, val.Error
		}
	}

	return g.pollingMgr.WaitForNotice(ctx, roomID, data, S{})
}

func (g *GoPolling) WaitForSelectedNotice(ctx context.Context, roomID string, data interface{}, selector S) (interface{}, error) {
	if g.buffer != nil {
		key := bufferKey{roomID, selector}
		hashedKey := getKeyHash(key)
		if val, ok := g.buffer.Find(hashedKey); ok {
			return val.Data, val.Error
		}
	}

	return g.pollingMgr.WaitForNotice(ctx, roomID, data, selector)
}

func (g *GoPolling) SubscribeListener(roomID string, lf ListenerFunc) {
	g.listenerMgr.Subscribe(roomID, lf)
}

func (g *GoPolling) Notify(roomID string, data interface{}, err error, selector S) error {
	msg := Message{data, err, selector}
	if g.buffer != nil {
		key := bufferKey{roomID, selector}
		hashedKey := getKeyHash(key)
		g.buffer.Save(hashedKey, msg, g.retention)
	}

	return g.bus.Publish(g.pubsubPrefix+roomID, msg)
}
