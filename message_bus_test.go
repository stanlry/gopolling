package gopolling

import (
	"sync"
	"testing"
	"time"
)

func TestGoroutineBus_EnqueueDequeue(t *testing.T) {
	bus := newGoroutineBus()

	var wg sync.WaitGroup
	room := "test_queue"
	data := "test"

	wg.Add(1)
	go func() {
		tick := time.Tick(1 * time.Second)
		select {
		case <-tick:
			t.Error("timeout")
			break
		case ev := <-bus.Dequeue(room):
			if ev.Data != data {
				t.Error("wrong event dat")
			}
			break
		}
		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)

	bus.Enqueue(room, Event{
		Data: data,
	})

	wg.Wait()
}

func TestGoroutineBus_PublishSubscribe(t *testing.T) {
	bus := newGoroutineBus()

	room := "test-room"
	data := "test data"

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			sub, err := bus.Subscribe(room)
			if err != nil {
				t.Error(err)
				wg.Done()
				return
			}

			if index%11 == 0 {
				bus.Unsubscribe(sub)
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

	if err := bus.Publish(room, Message{data, nil, S{}}); err != nil {
		t.Error(err)
	}

	wg.Wait()
}
