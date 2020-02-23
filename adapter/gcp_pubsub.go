package adapter

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/rs/xid"
	"github.com/stanlry/gopolling"
	"sync"
)

func NewGCPPubSubAdapter(client *pubsub.Client) *GCPPubSub {
	p := &GCPPubSub{
		client:      client,
		log:         &gopolling.NoOpLog{},
		subscribers: cmap.New(),
		listenMap:   make(map[string]context.CancelFunc),
	}
	p.start()
	return p
}

type GCPPubSub struct {
	client *pubsub.Client
	log    gopolling.Log

	subscribers cmap.ConcurrentMap
	listenMap   map[string]context.CancelFunc
	m           sync.RWMutex
}

func (g *GCPPubSub) start() {
	ctx, cf := context.WithCancel(context.Background())
	ctx2, cf2 := context.WithCancel(context.Background())
	g.m.Lock()
	g.listenMap[pubSubName] = cf
	g.listenMap[queueName] = cf2
	g.m.Unlock()
	go g.listenToSubscription(ctx, pubSubName, false, g.handleMessage)
	go g.listenToSubscription(ctx2, queueName, true, g.handleEvent)
}

func (g *GCPPubSub) Shutdown() {
	g.m.RLock()
	// invoke two cancel functions
	cf := g.listenMap[pubSubName]
	cf()
	cf = g.listenMap[queueName]
	cf()
	g.m.RUnlock()
}

func (g *GCPPubSub) getTopic(name string) (*pubsub.Topic, error) {
	ctx := context.Background()

	topic := g.client.Topic(name)
	exist, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if !exist {
		topic, err = g.client.CreateTopic(ctx, name)
		if err != nil {
			return nil, err
		}
	}
	return topic, nil
}

func (g *GCPPubSub) Publish(_ string, msg gopolling.Message) error {
	topic, err := g.getTopic(pubSubName)
	if err != nil {
		return err
	}
	defer topic.Stop()

	data, _ := json.Marshal(msg)
	ctx := context.Background()
	r := topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})
	_, err = r.Get(ctx)
	return err
}

func (g *GCPPubSub) handleMessage(ctx context.Context, m *pubsub.Message) {
	var msg gopolling.Message
	if err := json.Unmarshal(m.Data, &msg); err != nil {
		g.log.Errorf("fail to unmarshal message, error: %v", err)
		return
	}

	if val, ok := g.subscribers.Get(msg.Channel); ok {
		subMap := val.(*cmap.ConcurrentMap)
		subMap.IterCb(func(_ string, v interface{}) {
			ch := v.(chan gopolling.Message)
			ch <- msg
		})
	}

	m.Ack()
}

func (g *GCPPubSub) handleEvent(ctx context.Context, m *pubsub.Message) {
	var ev gopolling.Event
	if err := json.Unmarshal(m.Data, &ev); err != nil {
		g.log.Errorf("fail to unmarshal message, error: %v", err)
		return
	}

	if val, ok := g.subscribers.Get(ev.Channel); ok {
		ch := val.(chan gopolling.Event)
		ch <- ev
	}

	m.Ack()
}

type pubsubHandler func(context.Context, *pubsub.Message)

func (g *GCPPubSub) listenToSubscription(ctx context.Context, topicName string, exact bool, handler pubsubHandler) {
	subID := subPrefix + topicName
	if !exact {
		subID = subPrefix + xid.New().String()
	}

	sub := g.client.Subscription(subID)
	exist, err := sub.Exists(ctx)
	if err != nil {
		g.log.Errorf("fail to check subscription existence, error: %v", err)
		return
	}

	if !exist {
		topic, err := g.getTopic(topicName)
		if err != nil {
			g.log.Errorf("fail to get topic, error: %v", err)
			return
		}
		sub, err = g.client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{Topic: topic})
		if err != nil {
			g.log.Errorf("fail to create subscription, error: %v", err)
			return
		}

		if !exact {
			defer func() {
				if err := sub.Delete(ctx); err != nil {
					g.log.Errorf("fail to delete subscription, error: %v", err)
				}
			}()
		}
	}

	for {
		err := sub.Receive(ctx, handler)

		if err == context.Canceled {
			break
		}

		if err != nil {
			g.log.Errorf("error receiving from subscription %v, error :%v", subID, err)
			break
		}
	}
}

func (g *GCPPubSub) Subscribe(channel string) (gopolling.Subscription, error) {
	id := xid.New().String()
	ch := make(chan gopolling.Message)

	if val, ok := g.subscribers.Get(channel); ok {
		subMap := val.(*cmap.ConcurrentMap)
		subMap.Set(id, ch)
	} else {
		ss := cmap.New()
		ss.Set(id, ch)
		g.subscribers.Set(channel, &ss)
	}

	return gopolling.NewDefaultSubscription(channel, id, ch), nil
}

func (g *GCPPubSub) Unsubscribe(sub gopolling.Subscription) error {
	rsub := sub.(*gopolling.DefaultSubscription)
	if val, ok := g.subscribers.Get(rsub.Channel); ok {
		subMap := val.(*cmap.ConcurrentMap)
		subMap.Remove(rsub.ID)
	}
	return nil
}

func (g *GCPPubSub) Enqueue(_ string, event gopolling.Event) {
	topic, err := g.getTopic(queueName)
	if err != nil {
		g.log.Errorf("fail to get topic, error: %v", err)
		return
	}
	defer topic.Stop()

	data, _ := json.Marshal(event)
	ctx := context.Background()
	r := topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})
	if _, err = r.Get(ctx); err != nil {
		g.log.Errorf("fail deliver enqueue message, error: %v", err)
	}
}

func (g *GCPPubSub) Dequeue(channel string) <-chan gopolling.Event {
	ch := make(chan gopolling.Event)
	g.subscribers.Set(channel, ch)
	return ch
}

func (g *GCPPubSub) SetLog(log gopolling.Log) {
	g.log = log
}
