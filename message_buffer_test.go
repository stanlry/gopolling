package gopolling

import (
	"testing"
	"time"
)

func TestMemoryBuffer_SaveFind(t *testing.T) {
	buf := newMemoryBuffer()

	key := "tt"
	data := "test data"
	buf.Save(key, Message{key, data, nil, nil}, 10)
	val, ok := buf.Find(key)
	if !ok {
		t.Error("should be able to find the value")
	}
	if val.Data != data {
		t.Error("wrong data")
	}
}

func TestMemoryBuffer_Timeout(t *testing.T) {
	buf := newMemoryBuffer()

	key := "tt"
	data := "test data"
	buf.Save(key, Message{key, data, nil, nil}, 0)
	_, ok := buf.Find(key)
	if ok {
		t.Error("should not be able find")
	}

	buf.Save(key, Message{key, data, nil, nil}, 1)
	time.Sleep(1100 * time.Millisecond)
	_, ok = buf.Find(key)
	if ok {
		t.Error("should not be able find")
	}
}
