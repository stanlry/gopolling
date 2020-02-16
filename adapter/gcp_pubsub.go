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

const (
	subPrefix  = "gopolling_sub_"
	pubSubName = "gopolling_pubsub"
	queueName  = "gopolling_event_queue"
)

type elm struct{}

func newClientMap() *clientMap {
	return &clientMap{
		store: make(map[string]map[string]elm),
	}
}

type clientMap struct {
	store map[string]map[string]elm
	m     sync.RWMutex
}

func (q *clientMap) Add(key, id string) {
	q.m.Lock()
	if _, ok := q.store[key]; !ok {
		q.store[key] = make(map[string]elm)
	}
	q.store[key][id] = elm{}
	q.m.Unlock()
}

func (q *clientMap) Del(key, id string) {
	q.m.Lock()
	if _, ok := q.store[key]; ok {
		delete(q.store[key], id)
	}
	q.m.Unlock()
}

func (q *clientMap) Get(key string) []string {
	var rlist []string

	q.m.RLock()
	if cm, ok := q.store[key]; ok {
		rlist = make([]string, len(cm))
		count := 0
		for i := range cm {
			rlist[count] = i
			count++
		}
	}
	q.m.RUnlock()

	return rlist
}

type gcpSubscription struct {
	id   string
	room string
	ch   chan gopolling.Message
}

func (g *gcpSubscription) Receive() <-chan gopolling.Message {
	return g.ch
}

func NewGCPPubSubAdapter(client *pubsub.Client) *GCPPubSub {
	p := &GCPPubSub{
		client:    client,
		log:       &gopolling.NoOpLog{},
		cm:        newClientMap(),
		chanMap:   cmap.New(),
		listenMap: make(map[string]context.CancelFunc),
	}
	p.start()
	return p
}

type GCPPubSub struct {
	client *pubsub.Client
	log    gopolling.Log

	cm        *clientMap
	chanMap   cmap.ConcurrentMap
	listenMap map[string]context.CancelFunc
	m         sync.RWMutex
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

	chanIDs := g.cm.Get(msg.Channel)
	for _, id := range chanIDs {
		if val, ok := g.chanMap.Get(id); ok {
			ch := val.(chan gopolling.Message)
			ch <- msg
		}
	}

	m.Ack()
}

func (g *GCPPubSub) handleEvent(ctx context.Context, m *pubsub.Message) {
	var ev gopolling.Event
	if err := json.Unmarshal(m.Data, &ev); err != nil {
		g.log.Errorf("fail to unmarshal message, error: %v", err)
		return
	}

	chanIDs := g.cm.Get(ev.Channel)
	for _, id := range chanIDs {
		if val, ok := g.chanMap.Get(id); ok {
			ch := val.(chan gopolling.Event)
			ch <- ev
		}
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

func (g *GCPPubSub) Subscribe(room string) (gopolling.Subscription, error) {
	id := xid.New().String()
	ch := make(chan gopolling.Message)
	gSub := gcpSubscription{
		id:   id,
		room: room,
		ch:   ch,
	}

	g.chanMap.Set(id, ch)
	g.cm.Add(room, id)

	return &gSub, nil
}

func (g *GCPPubSub) Unsubscribe(sub gopolling.Subscription) error {
	gSub := sub.(*gcpSubscription)
	g.cm.Del(gSub.room, gSub.id)
	g.chanMap.Remove(gSub.id)
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
	id := xid.New().String()
	ch := make(chan gopolling.Event)
	g.chanMap.Set(id, ch)
	g.cm.Add(channel, id)

	return ch
}

func (g *GCPPubSub) SetLog(log gopolling.Log) {
	g.log = log
}
