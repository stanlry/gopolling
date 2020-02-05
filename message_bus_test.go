package gopolling

import (
	"sync"
	"testing"
	"time"
)

func TestGoroutineBus_EnqueueDequeue(t *testing.T) {
	bus := newGoroutineBus()

	var wg sync.WaitGroup
	room := "testroom"
	data := "test"
	go func() {
		wg.Add(1)
		select {
		case ev := <-bus.Dequeue(room):
			if ev.Data != data {
				t.Error("wrong event dat")
			}
			break
		}
		wg.Done()
	}()

	bus.Enqueue(room, Event{
		Data: data,
	})

	wg.Wait()
}

func TestGoroutineBus_PublishSubscribe(t *testing.T) {
	bus := newGoroutineBus()

	room := "testroom"
	data := "test data"

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			sub, err := bus.Subscribe(room)
			if err != nil {
				t.Error(err)
			}

			if index == 4 {
				sub.Unsubscribe()
				wg.Done()
				return
			}

			tick := time.Tick(1 * time.Second)
			select {
			case <-tick:
				t.Error("cannot receive message")
			case msg := <-sub.Receive():
				if msg.Data != data {
					t.Error("invalid data")
				}
				wg.Done()
			}
		}(i)
	}

	time.Sleep(1 * time.Millisecond)

	if err := bus.Publish(room, Message{Data: data}); err != nil {
		t.Error(err)
	}

	wg.Wait()
}
