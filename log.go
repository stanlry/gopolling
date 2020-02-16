package gopolling

// Logger interface support error logging
type Log interface {
	Errorf(string, ...interface{})
}

// No Operation Logger
type NoOpLog struct{}

// Log error but no action will be taken
func (n *NoOpLog) Errorf(st string, args ...interface{}) {}

type Loggable interface {
	SetLog(Log)
}
