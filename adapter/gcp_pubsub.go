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
	return &GCPPubSub{
		client:    client,
		log:       &gopolling.NoOpLog{},
		cm:        newClientMap(),
		chanMap:   cmap.New(),
		listenMap: make(map[string]context.CancelFunc),
	}
}

type GCPPubSub struct {
	client *pubsub.Client
	log    gopolling.Log

	cm        *clientMap
	chanMap   cmap.ConcurrentMap
	listenMap map[string]context.CancelFunc
	m         sync.RWMutex
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

func (g *GCPPubSub) Publish(room string, msg gopolling.Message) error {
	topic, err := g.getTopic(room)
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

func (g *GCPPubSub) listenToSubscription(ctx context.Context, tp string) {
	sub := g.client.Subscription(tp)
	exist, err := sub.Exists(ctx)
	if err != nil {
		g.log.Errorf("fail to check subscription existence, error: %v", err)
		return
	}

	if !exist {
		topic, err := g.getTopic(tp)
		if err != nil {
			g.log.Errorf("fail to get topic, error: %v", err)
			return
		}
		sub, err = g.client.CreateSubscription(ctx, tp, pubsub.SubscriptionConfig{Topic: topic})
		if err != nil {
			g.log.Errorf("fail to create subscription, error: %v", err)
			return
		}
	}

	for {
		err := sub.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
			var msg gopolling.Message
			if err := json.Unmarshal(message.Data, &msg); err != nil {
				g.log.Errorf("fail to unmarshal message on subscript: %v, error: %v", sub.ID(), err)
				return
			}

			chanIDs := g.cm.Get(tp)
			for _, id := range chanIDs {
				if val, ok := g.chanMap.Get(id); ok {
					ch := val.(chan gopolling.Message)
					ch <- msg
				}
			}

			message.Ack()
		})

		if err == context.Canceled {
			return
		}

		if err != nil {
			g.log.Errorf("error receiving from subscription %v, error :%v", tp, err)
			return
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

	// check if subscription exists
	g.m.RLock()
	_, hasListen := g.listenMap[room]
	g.m.RUnlock()

	if !hasListen {
		g.m.Lock()
		ctx, cf := context.WithCancel(context.Background())
		g.listenMap[room] = cf
		go g.listenToSubscription(ctx, room)
		g.m.Unlock()
	}

	return &gSub, nil
}

func (g *GCPPubSub) Unsubscribe(sub gopolling.Subscription) error {
	gSub := sub.(*gcpSubscription)
	g.cm.Del(gSub.room, gSub.id)
	g.chanMap.Remove(gSub.id)
	return nil
}

func (g *GCPPubSub) Enqueue(string, gopolling.Event) {
	panic("implement me")
}

func (g *GCPPubSub) Dequeue(string) <-chan gopolling.Event {
	panic("implement me")
}

func (g *GCPPubSub) SetLog(log gopolling.Log) {
	g.log = log
}
