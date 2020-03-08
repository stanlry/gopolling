package gopolling

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPollingWithListener(t *testing.T) {
	mgr := New(DefaultOption)
	channel := "test-listener"
	senderData := "send"
	receivedData := "receive"
	mgr.SubscribeListener(channel, func(ev Event, cb *Callback) {
		if ev.Data != senderData {
			t.Error("invalid send data")
		}
		cb.Reply(receivedData)
	})
	time.Sleep(1 * time.Microsecond)

	val, err := mgr.WaitForNotice(context.TODO(), channel, senderData)
	if val != receivedData || err != nil {
		t.Errorf("invalid return: val: %v, err: %v", val, err)
	}
}

func TestPollingWithNotifier(t *testing.T) {
	mgr := New(DefaultOption)
	channel := "test"
	data := "haha data"

	go func() {
		time.Sleep(10 * time.Millisecond)
		if err := mgr.Notify(channel, data, S{"name": "123"}); err != nil {
			t.Error(err)
			return
		}
		if err := mgr.Notify(channel, data, S{"name": "567"}); err != nil {
			t.Error(err)
			return
		}
	}()

	val, err := mgr.WaitForSelectedNotice(context.TODO(), channel, data, S{"name": "567"})
	if val != data || err != nil {
		t.Errorf("invalid return: val: %v, err: %v", val, err)
	}
}

const ParallelLimit = 10

func TestPubPolling(t *testing.T) {
	mgr := New(Option{Timeout: 2 * time.Second})
	var wg sync.WaitGroup
	channel := "test3"
	data := "good data"

	for i := 0; i < ParallelLimit; i++ {
		wg.Add(1)
		go func() {
			val, err := mgr.WaitForNotice(context.TODO(), channel, nil)
			if val != data || err != nil {
				t.Errorf("invalid return: val: %v, err: %v", val, err)
			}
			wg.Done()
		}()
	}

	time.Sleep(100 * time.Millisecond)
	if err := mgr.Notify(channel, data, S{}); err != nil {
		t.Error(err)
		return
	}

	wg.Wait()
}
