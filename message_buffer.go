package gopolling

import (
	"encoding/hex"
	"github.com/orcaman/concurrent-map"
	"hash/fnv"
	"strings"
	"time"
)

type MessageBuffer interface {
	Find(string) (Message, bool)
	Save(string, Message, int)
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

func newMemoryBuffer() MessageBuffer {
	return &memoryBuffer{buffer: cmap.New()}
}

type bufElm struct {
	Message Message
	Timeout time.Time
}

type memoryBuffer struct {
	buffer cmap.ConcurrentMap
}

func (m *memoryBuffer) Find(key string) (Message, bool) {
	if val, ok := m.buffer.Get(key); ok {
		el := val.(bufElm)
		if time.Now().After(el.Timeout) {
			return Message{}, false
		}

		return el.Message, true
	}

	return Message{}, false
}

func (m *memoryBuffer) Save(key string, msg Message, t int) {
	if t == 0 {
		return
	}

	timeout := time.Second * time.Duration(t)
	el := bufElm{msg, time.Now().Add(timeout)}
	m.buffer.Set(key, el)
}
