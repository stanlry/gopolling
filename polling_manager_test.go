package gopolling

import (
	"context"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

func setupManager(mc *gomock.Controller) (*PollingManager, *MockSubscription) {
	mockBus := NewMockMessageBus(mc)
	mockSub := NewMockSubscription(mc)

	mockBus.EXPECT().Subscribe(gomock.Any()).Return(mockSub, nil).Times(1)
	mockBus.EXPECT().Enqueue(gomock.Any(), gomock.Any()).Times(1)
	mgr := NewPollingManager(mockBus, 10*time.Millisecond, queuePrefix, pubsubPrefix)

	return &mgr, mockSub
}

func TestPollingManager_Timeout(t *testing.T) {
	mc := gomock.NewController(t)
	defer mc.Finish()

	mgr, mockSub := setupManager(mc)
	mockSub.EXPECT().Receive().Return(make(chan Message))
	mockSub.EXPECT().Unsubscribe().Times(1)

	mgr.timeout = 10 * time.Millisecond

	val, err := mgr.WaitForNotice(context.TODO(), "test", nil, S{})
	if val != nil || err != ErrTimeout {
		t.Errorf("polling client should timeout on no reply")
	}
}

func TestPollingManager_CancelContext(t *testing.T) {
	mc := gomock.NewController(t)
	defer mc.Finish()

	mgr, mockSub := setupManager(mc)
	mockSub.EXPECT().Receive().Return(make(chan Message))
	mockSub.EXPECT().Unsubscribe().Times(1)

	ctx, cf := context.WithCancel(context.Background())

	go func() {
		time.Sleep(1 * time.Millisecond)
		cf()
	}()

	val, err := mgr.WaitForNotice(ctx, "test", nil, S{})
	if val != nil || err != ErrCancelled {
		t.Errorf("should reutrn cancelled on context cancel")
	}
}

func TestPollingManager_Selector(t *testing.T) {
	mc := gomock.NewController(t)
	defer mc.Finish()

	sel := S{"name": "s1"}
	data := "test data"

	mgr, mockSub := setupManager(mc)
	ch := make(chan Message, 2)
	ch <- Message{Data: data}
	ch <- Message{Data: data, Selector: sel}
	mockSub.EXPECT().Receive().Return(ch).Times(2)
	mockSub.EXPECT().Unsubscribe().Times(1)

	val, err := mgr.WaitForNotice(context.TODO(), "test", nil, sel)
	if val != data || err != nil {
		t.Errorf("should have received the message")
	}
}

func TestPollingManager_ReplyID(t *testing.T) {
	mc := gomock.NewController(t)
	defer mc.Finish()

	data := "test data"
	ch := make(chan Message, 2)
	room := "test"

	mockBus := NewMockMessageBus(mc)
	mockSub := NewMockSubscription(mc)

	mgr := NewPollingManager(mockBus, 10*time.Millisecond, queuePrefix, pubsubPrefix)
	mockBus.EXPECT().Subscribe(pubsubPrefix+room).Return(mockSub, nil).Times(1)
	mockSub.EXPECT().Receive().Return(ch).Times(2)
	mockSub.EXPECT().Unsubscribe().Times(1)
	mockBus.EXPECT().Enqueue(queuePrefix+room, gomock.Any()).Do(func(roomID string, msg Event) {
		ch <- Message{Data: data, Selector: S{idKey: msg.Selector[idKey]}}
	})
	ch <- Message{Data: "fake data", Selector: S{idKey: "fake id"}}

	val, err := mgr.WaitForNotice(context.TODO(), room, nil, S{})
	if val != data || err != nil {
		t.Errorf("should have received the message")
	}
}

func TestCompareSelector(t *testing.T) {
	a := S{"name": "123"}
	b := S{"name": "123"}

	if compareSelectors(a, b) != true {
		t.Errorf("selectors should be equal")
	}

	// different field number
	b["tel"] = "123"
	if compareSelectors(a, b) != false {
		t.Errorf("selectors is not equal")
	}

	// different field value
	b = S{"name": "haha"}
	if compareSelectors(a, b) != false {
		t.Errorf("selectors is not equal")
	}
}
