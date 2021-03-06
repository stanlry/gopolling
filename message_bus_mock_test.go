// Code generated by MockGen. DO NOT EDIT.
// Source: message_bus.go

// Package gopolling is a generated GoMock package.
package gopolling

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockSubscription is a mock of Subscription interface
type MockSubscription struct {
	ctrl     *gomock.Controller
	recorder *MockSubscriptionMockRecorder
}

// MockSubscriptionMockRecorder is the mock recorder for MockSubscription
type MockSubscriptionMockRecorder struct {
	mock *MockSubscription
}

// NewMockSubscription creates a new mock instance
func NewMockSubscription(ctrl *gomock.Controller) *MockSubscription {
	mock := &MockSubscription{ctrl: ctrl}
	mock.recorder = &MockSubscriptionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSubscription) EXPECT() *MockSubscriptionMockRecorder {
	return m.recorder
}

// Receive mocks base method
func (m *MockSubscription) Receive() <-chan Message {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Receive")
	ret0, _ := ret[0].(<-chan Message)
	return ret0
}

// Receive indicates an expected call of Receive
func (mr *MockSubscriptionMockRecorder) Receive() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Receive", reflect.TypeOf((*MockSubscription)(nil).Receive))
}

// MockPubSub is a mock of PubSub interface
type MockPubSub struct {
	ctrl     *gomock.Controller
	recorder *MockPubSubMockRecorder
}

// MockPubSubMockRecorder is the mock recorder for MockPubSub
type MockPubSubMockRecorder struct {
	mock *MockPubSub
}

// NewMockPubSub creates a new mock instance
func NewMockPubSub(ctrl *gomock.Controller) *MockPubSub {
	mock := &MockPubSub{ctrl: ctrl}
	mock.recorder = &MockPubSubMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPubSub) EXPECT() *MockPubSubMockRecorder {
	return m.recorder
}

// Publish mocks base method
func (m *MockPubSub) Publish(arg0 string, arg1 Message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Publish", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Publish indicates an expected call of Publish
func (mr *MockPubSubMockRecorder) Publish(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockPubSub)(nil).Publish), arg0, arg1)
}

// Subscribe mocks base method
func (m *MockPubSub) Subscribe(arg0 string) (Subscription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe", arg0)
	ret0, _ := ret[0].(Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Subscribe indicates an expected call of Subscribe
func (mr *MockPubSubMockRecorder) Subscribe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockPubSub)(nil).Subscribe), arg0)
}

// Unsubscribe mocks base method
func (m *MockPubSub) Unsubscribe(arg0 Subscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unsubscribe", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unsubscribe indicates an expected call of Unsubscribe
func (mr *MockPubSubMockRecorder) Unsubscribe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unsubscribe", reflect.TypeOf((*MockPubSub)(nil).Unsubscribe), arg0)
}

// MockEventQueue is a mock of EventQueue interface
type MockEventQueue struct {
	ctrl     *gomock.Controller
	recorder *MockEventQueueMockRecorder
}

// MockEventQueueMockRecorder is the mock recorder for MockEventQueue
type MockEventQueueMockRecorder struct {
	mock *MockEventQueue
}

// NewMockEventQueue creates a new mock instance
func NewMockEventQueue(ctrl *gomock.Controller) *MockEventQueue {
	mock := &MockEventQueue{ctrl: ctrl}
	mock.recorder = &MockEventQueueMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEventQueue) EXPECT() *MockEventQueueMockRecorder {
	return m.recorder
}

// Enqueue mocks base method
func (m *MockEventQueue) Enqueue(arg0 string, arg1 Event) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Enqueue", arg0, arg1)
}

// Enqueue indicates an expected call of Enqueue
func (mr *MockEventQueueMockRecorder) Enqueue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enqueue", reflect.TypeOf((*MockEventQueue)(nil).Enqueue), arg0, arg1)
}

// Dequeue mocks base method
func (m *MockEventQueue) Dequeue(arg0 string) <-chan Event {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dequeue", arg0)
	ret0, _ := ret[0].(<-chan Event)
	return ret0
}

// Dequeue indicates an expected call of Dequeue
func (mr *MockEventQueueMockRecorder) Dequeue(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dequeue", reflect.TypeOf((*MockEventQueue)(nil).Dequeue), arg0)
}

// MockLoggable is a mock of Loggable interface
type MockLoggable struct {
	ctrl     *gomock.Controller
	recorder *MockLoggableMockRecorder
}

// MockLoggableMockRecorder is the mock recorder for MockLoggable
type MockLoggableMockRecorder struct {
	mock *MockLoggable
}

// NewMockLoggable creates a new mock instance
func NewMockLoggable(ctrl *gomock.Controller) *MockLoggable {
	mock := &MockLoggable{ctrl: ctrl}
	mock.recorder = &MockLoggableMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLoggable) EXPECT() *MockLoggableMockRecorder {
	return m.recorder
}

// SetLog mocks base method
func (m *MockLoggable) SetLog(arg0 Log) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetLog", arg0)
}

// SetLog indicates an expected call of SetLog
func (mr *MockLoggableMockRecorder) SetLog(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetLog", reflect.TypeOf((*MockLoggable)(nil).SetLog), arg0)
}

// MockMessageBus is a mock of MessageBus interface
type MockMessageBus struct {
	ctrl     *gomock.Controller
	recorder *MockMessageBusMockRecorder
}

// MockMessageBusMockRecorder is the mock recorder for MockMessageBus
type MockMessageBusMockRecorder struct {
	mock *MockMessageBus
}

// NewMockMessageBus creates a new mock instance
func NewMockMessageBus(ctrl *gomock.Controller) *MockMessageBus {
	mock := &MockMessageBus{ctrl: ctrl}
	mock.recorder = &MockMessageBusMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMessageBus) EXPECT() *MockMessageBusMockRecorder {
	return m.recorder
}

// Publish mocks base method
func (m *MockMessageBus) Publish(arg0 string, arg1 Message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Publish", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Publish indicates an expected call of Publish
func (mr *MockMessageBusMockRecorder) Publish(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockMessageBus)(nil).Publish), arg0, arg1)
}

// Subscribe mocks base method
func (m *MockMessageBus) Subscribe(arg0 string) (Subscription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe", arg0)
	ret0, _ := ret[0].(Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Subscribe indicates an expected call of Subscribe
func (mr *MockMessageBusMockRecorder) Subscribe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockMessageBus)(nil).Subscribe), arg0)
}

// Unsubscribe mocks base method
func (m *MockMessageBus) Unsubscribe(arg0 Subscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unsubscribe", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unsubscribe indicates an expected call of Unsubscribe
func (mr *MockMessageBusMockRecorder) Unsubscribe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unsubscribe", reflect.TypeOf((*MockMessageBus)(nil).Unsubscribe), arg0)
}

// Enqueue mocks base method
func (m *MockMessageBus) Enqueue(arg0 string, arg1 Event) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Enqueue", arg0, arg1)
}

// Enqueue indicates an expected call of Enqueue
func (mr *MockMessageBusMockRecorder) Enqueue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enqueue", reflect.TypeOf((*MockMessageBus)(nil).Enqueue), arg0, arg1)
}

// Dequeue mocks base method
func (m *MockMessageBus) Dequeue(arg0 string) <-chan Event {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dequeue", arg0)
	ret0, _ := ret[0].(<-chan Event)
	return ret0
}

// Dequeue indicates an expected call of Dequeue
func (mr *MockMessageBusMockRecorder) Dequeue(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dequeue", reflect.TypeOf((*MockMessageBus)(nil).Dequeue), arg0)
}

// SetLog mocks base method
func (m *MockMessageBus) SetLog(arg0 Log) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetLog", arg0)
}

// SetLog indicates an expected call of SetLog
func (mr *MockMessageBusMockRecorder) SetLog(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetLog", reflect.TypeOf((*MockMessageBus)(nil).SetLog), arg0)
}
