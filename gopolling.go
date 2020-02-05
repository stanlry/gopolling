package gopolling

import (
	"context"
	"time"
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
}

var DefaultOption = Option{
	Retention: 120,
	Timeout:   120 * time.Second,
	Bus:       newGoroutineBus(),
	Buffer:    newMemoryBuffer(),
}

func New(option Option) GoPolling {
	var gp GoPolling

	if option.Bus != nil {
		gp.bus = option.Bus
	} else {
		panic("no message bus")
	}

	gp.pollingMgr = NewPollingManager(option.Bus)
	if option.Timeout != 0 {
		gp.pollingMgr.timeout = option.Timeout
	}

	gp.listenerMgr = NewListenerManager(option.Bus)

	if option.Logger != nil {
		gp.pollingMgr.log = option.Logger
		gp.listenerMgr.log = option.Logger
		gp.bus.SetLog(option.Logger)
	}

	if option.Buffer != nil {
		gp.buffer = option.Buffer
	}

	if option.Retention != 0 {
		gp.retention = option.Retention
	}

	return gp
}

type GoPolling struct {
	bus         MessageBus
	pollingMgr  PollingManager
	listenerMgr ListenerManager
	buffer      MessageBuffer
	retention   int
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

func (g *GoPolling) Notify(roomID string, message Message) error {
	if g.buffer != nil {
		key := bufferKey{roomID, message.Selector}
		hashedKey := getKeyHash(key)
		g.buffer.Save(hashedKey, message, g.retention)
	}

	return g.bus.Publish(roomID, message)
}
