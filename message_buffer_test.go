package gopolling

import (
	"testing"
	"time"
)

func TestMemoryBuffer_SaveFind(t *testing.T) {
	buf := newMemoryBuffer()

	key := "tt"
	data := "test data"
	buf.Save(key, data, nil, 10)
	val, ok := buf.Find(key)
	if !ok {
		t.Error("should be able to find the value")
	}
	if val.Payload.Data() != data {
		t.Error("wrong data")
	}
}

func TestMemoryBuffer_Timeout(t *testing.T) {
	buf := newMemoryBuffer()

	key := "tt"
	data := "test data"
	buf.Save(key, data, nil, 0)
	_, ok := buf.Find(key)
	if ok {
		t.Error("should not be able find")
	}

	buf.Save(key, data, nil, 1)
	time.Sleep(1100 * time.Millisecond)
	_, ok = buf.Find(key)
	if ok {
		t.Error("should not be able find")
	}
}
