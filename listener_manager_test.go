package gopolling

import (
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

func setupListener(mc *gomock.Controller) (*ListenerManager, *MockMessageBus) {
	bus := NewMockMessageBus(mc)
	mgr := NewListenerManager(bus)

	return &mgr, bus
}

func noOpFunc(Event, *Callback) {}

func TestListenerManager_Subscribe(t *testing.T) {
	mc := gomock.NewController(t)
	defer mc.Finish()

	mgr, bus := setupListener(mc)

	room := "room1"
	bus.EXPECT().Dequeue(room).Times(1)
	mgr.Subscribe(room, noOpFunc)

	time.Sleep(10 * time.Millisecond)
}

func TestListenerManager_Notify(t *testing.T) {
	mc := gomock.NewController(t)
	defer mc.Finish()

	mgr, bus := setupListener(mc)

	room := "room1"
	notifyData := "testNotify"
	ch := make(chan Event)
	bus.EXPECT().Dequeue(room).Return(ch).Times(2)
	bus.EXPECT().Publish(room, gomock.Eq(Message{
		Data: notifyData,
	})).Times(1)

	mgr.Subscribe(room, func(event Event, callback *Callback) {
		callback.Reply(notifyData, nil)
	})

	ch <- Event{}

	time.Sleep(10 * time.Millisecond)
}
