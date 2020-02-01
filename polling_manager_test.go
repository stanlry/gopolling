package gopolling

import (
	"context"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

func setupManager(mc *gomock.Controller) (*PollingManager, *MockSubscription) {
	mockAdapter := NewMockMessageBus(mc)
	mockSub := NewMockSubscription(mc)

	mockAdapter.EXPECT().Subscribe(gomock.Any()).Return(mockSub, nil).Times(1)
	mockAdapter.EXPECT().Enqueue(gomock.Any(), gomock.Any()).Times(1)
	mgr := NewPollingManager(mockAdapter)

	return &mgr, mockSub
}

func TestPollingManagerTimeout(t *testing.T) {
	mc := gomock.NewController(t)
	defer mc.Finish()

	mgr, mockSub := setupManager(mc)
	mockSub.EXPECT().Receive().Return(make(chan Message))
	mockSub.EXPECT().Unsubscribe().Times(1)

	mgr.timeout = 1 * time.Second

	val, err := mgr.WaitForNotice(context.TODO(), "test", nil, S{})
	if val != nil || err != ErrTimeout {
		t.Errorf("polling client should timeout on no reply")
	}
}

func TestPollingManagerCancelContext(t *testing.T) {
	mc := gomock.NewController(t)
	defer mc.Finish()

	mgr, mockSub := setupManager(mc)
	mockSub.EXPECT().Receive().Return(make(chan Message))
	mockSub.EXPECT().Unsubscribe().Times(1)

	ctx, cf := context.WithCancel(context.Background())

	go func() {
		time.Sleep(10 * time.Millisecond)
		cf()
	}()

	val, err := mgr.WaitForNotice(ctx, "test", nil, S{})
	if val != nil || err != ErrCancelled {
		t.Errorf("should reutrn cancelled on context cancel")
	}
}

func TestPollingManagerSelector(t *testing.T) {
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
