package gopolling

import (
	"encoding/hex"
	"errors"
	"github.com/orcaman/concurrent-map"
	"hash/fnv"
	"reflect"
	"strings"
	"time"
)

type MessageBuffer interface {
	Find(string) (Msg, bool)
	Save(string, interface{}, error, int)
}

func getKeyHash(k bufferKey) string {
	var st strings.Builder
	st.WriteString(k.RoomID)
	for k, v := range k.Selector {
		st.WriteString(k)
		st.WriteString(v)
	}

	hfunc := fnv.New32a()
	hfunc.Write([]byte(st.String()))
	return hex.EncodeToString(hfunc.Sum(nil))
}

type bufferKey struct {
	RoomID   string
	Selector S
}

func newMemoryPayload(data interface{}) Payload {
	return &memoryPayload{data}
}

type memoryPayload struct {
	data interface{}
}

func (m *memoryPayload) Data() interface{} {
	return m.data
}

func (m *memoryPayload) Decode(t interface{}) error {
	if reflect.TypeOf(t).Kind() != reflect.Ptr {
		return errors.New("not a pointer")
	}

	switch reflect.TypeOf(m.data).Kind() {
	case reflect.Ptr:
		t = m.data
	default:
		t = reflect.ValueOf(m.data).Pointer()
	}
	return nil
}

func newMemoryBuffer() MessageBuffer {
	return &memoryBuffer{buffer: cmap.New()}
}

type bufElm struct {
	Message Msg
	Timeout time.Time
}

type memoryBuffer struct {
	buffer cmap.ConcurrentMap
}

func (m *memoryBuffer) Find(key string) (Msg, bool) {
	if val, ok := m.buffer.Get(key); ok {
		el := val.(bufElm)
		if time.Now().After(el.Timeout) {
			return Msg{}, false
		}

		return el.Message, true
	}

	return Msg{}, false
}

func (m *memoryBuffer) Save(key string, data interface{}, err error, t int) {
	if t == 0 {
		return
	}

	timeout := time.Second * time.Duration(t)
	el := bufElm{Msg{newMemoryPayload(data), err, nil}, time.Now().Add(timeout)}
	m.buffer.Set(key, el)
}
