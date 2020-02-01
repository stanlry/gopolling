package gopolling

type Log interface {
	Errorf(string, ...interface{})
}

type NoOpLog struct{}

func (n *NoOpLog) Errorf(string, ...interface{}) {}
