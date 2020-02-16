package gopolling

import (
	"sync"
	"testing"
	"time"
)

func TestGoroutineBus_EnqueueDequeue(t *testing.T) {
	bus := newGoroutineBus()

	var wg sync.WaitGroup
	channel := "test_queue"
	data := "test"

	wg.Add(1)
	go func() {
		tick := time.Tick(1 * time.Second)
		select {
		case <-tick:
			t.Error("timeout")
			break
		case ev := <-bus.Dequeue(channel):
			if ev.Data != data {
				t.Error("wrong event dat")
			}
			break
		}
		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)

	bus.Enqueue(channel, Event{
		Data: data,
	})

	wg.Wait()
}

func TestGoroutineBus_PublishSubscribe(t *testing.T) {
	bus := newGoroutineBus()

	channel := "test-channel"
	data := "test data"

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			sub, err := bus.Subscribe(channel)
			if err != nil {
				t.Error(err)
				wg.Done()
				return
			}

			if index%11 == 0 {
				if err := bus.Unsubscribe(sub); err != nil {
					t.Error(err)
				}
				wg.Done()
				return
			}

			tick := time.Tick(1 * time.Second)
			select {
			case <-tick:
				t.Error("cannot receive message")
				wg.Done()
			case msg := <-sub.Receive():
				if msg.Data != data {
					t.Error("invalid data")
				}
				wg.Done()
			}
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	if err := bus.Publish(channel, Message{channel, data, nil, S{}}); err != nil {
		t.Error(err)
	}

	wg.Wait()
}
