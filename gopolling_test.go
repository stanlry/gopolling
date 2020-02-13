package gopolling

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPollingWithListener(t *testing.T) {
	mgr := New(DefaultOption)
	roomID := "test-listener"
	senderData := "send"
	receivedData := "receive"
	mgr.SubscribeListener(roomID, func(ev Event, cb *Callback) {
		if ev.Data != senderData {
			t.Error("invalid send data")
		}
		cb.Reply(receivedData, nil)
	})
	time.Sleep(1 * time.Microsecond)

	val, err := mgr.WaitForNotice(context.TODO(), roomID, senderData)
	if val != receivedData || err != nil {
		t.Errorf("invalid return: val: %v, err: %v", val, err)
	}
}

func TestPollingWithNotifier(t *testing.T) {
	mgr := New(DefaultOption)
	roomID := "test"
	data := "haha data"

	go func() {
		time.Sleep(10 * time.Millisecond)
		mgr.Notify(roomID, data, nil, S{"name": "123"})
		mgr.Notify(roomID, data, nil, S{"name": "567"})
	}()

	val, err := mgr.WaitForSelectedNotice(context.TODO(), roomID, data, S{"name": "567"})
	if val != data || err != nil {
		t.Errorf("invalid return: val: %v, err: %v", val, err)
	}
}

const ParallelLimit = 10

func TestPubPolling(t *testing.T) {
	mgr := New(DefaultOption)
	var wg sync.WaitGroup
	room := "test3"
	data := "good data"

	for i := 0; i < ParallelLimit; i++ {
		wg.Add(1)
		go func() {
			val, err := mgr.WaitForNotice(context.TODO(), room, nil)
			if val != data || err != nil {
				t.Errorf("invalid return: val: %v, err: %v", val, err)
			}
			wg.Done()
		}()
	}

	time.Sleep(100 * time.Millisecond)
	mgr.Notify(room, data, nil, S{})

	wg.Wait()
}
