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

	// Message and task queue
	Adapter MessageAdapter

	// Logging implementation
	Logger Log
}

func NewGoPolling(option Option) GoPolling {
	var adapter MessageAdapter

	if option.Adapter != nil {
		adapter = option.Adapter
	} else {
		adapter = NewGoroutineAdapter()
	}

	pollingMgr := NewPollingManager(adapter)
	if option.Timeout != 0 {
		pollingMgr.timeout = option.Timeout
	}

	listenerMgr := NewListenerManager(adapter)

	if option.Logger != nil {
		pollingMgr.log = option.Logger
		listenerMgr.log = option.Logger
		adapter.SetLog(option.Logger)
	}

	return GoPolling{
		bus:         adapter,
		pollingMgr:  pollingMgr,
		listenerMgr: listenerMgr,
	}
}

type DataOption struct {
	Data     interface{}
	Selector S
}

type GoPolling struct {
	bus         MessageAdapter
	pollingMgr  PollingManager
	listenerMgr ListenerManager
}

func (g *GoPolling) WaitForNotice(ctx context.Context, roomID string, opt *DataOption) (interface{}, error) {
	if opt != nil {
		return g.pollingMgr.WaitForNotice(ctx, roomID, opt.Data, opt.Selector)
	} else {
		return g.pollingMgr.WaitForNotice(ctx, roomID, nil, S{})
	}
}

func (g *GoPolling) SubscribeListener(roomID string, lf ListenerFunc) {
	g.listenerMgr.Subscribe(roomID, lf)
}

func (g *GoPolling) Notify(roomID string, message Message) error {
	return g.bus.Publish(roomID, message)
}
