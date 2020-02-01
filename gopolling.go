package gopolling

import (
	"context"
	"time"
)

type Option struct {
	// Retention period indicate the TTL time for the notify message
	Retention time.Duration

	// Long polling client timeout (default 120s)
	Timeout time.Duration

	// Message Bus
	Bus MessageBus

	// Logging implementation
	Logger Log
}

func NewGoPolling(option Option) GoPolling {
	var bus MessageBus

	if option.Bus != nil {
		bus = option.Bus
	} else {
		panic("message bus not found")
	}

	pollingMgr := NewPollingManager(bus)
	if option.Timeout != 0 {
		pollingMgr.timeout = option.Timeout
	}

	listenerMgr := NewListenerManager(bus)

	if option.Logger != nil {
		pollingMgr.log = option.Logger
		listenerMgr.log = option.Logger
		bus.SetLog(option.Logger)
	}

	return GoPolling{
		bus:         bus,
		pollingMgr:  pollingMgr,
		listenerMgr: listenerMgr,
	}
}

type GoPolling struct {
	bus         MessageBus
	pollingMgr  PollingManager
	listenerMgr ListenerManager
}

func (g *GoPolling) WaitForNotice(ctx context.Context, roomID string, data interface{}) (interface{}, error) {
	return g.pollingMgr.WaitForNotice(ctx, roomID, data, S{})
}

func (g *GoPolling) WaitForSelectedNotice(ctx context.Context, roomID string, data interface{}, selector S) (interface{}, error) {
	return g.pollingMgr.WaitForNotice(ctx, roomID, data, selector)
}

func (g *GoPolling) SubscribeListener(roomID string, lf ListenerFunc) {
	g.listenerMgr.Subscribe(roomID, lf)
}

func (g *GoPolling) Notify(roomID string, message Message) error {
	return g.bus.Publish(roomID, message)
}
