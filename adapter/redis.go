package adapter

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/rs/xid"
	"github.com/stanlry/gopolling"
	"time"
)

func NewRedisAdapter(uri, password string) *RedisAdapter {
	ad := RedisAdapter{
		pubsubName:  pubSubName,
		subscribers: cmap.New(),
		log:         &gopolling.NoOpLog{},
	}

	ad.pool = &redis.Pool{
		MaxIdle:     1,
		MaxActive:   100,
		IdleTimeout: 5 * time.Minute,
		Dial: func() (redis.Conn, error) {
			option := redis.DialPassword(password)
			con, err := redis.Dial("tcp", uri, option)
			if err != nil {
				ad.log.Errorf("fail to connect to redis, error: %v", err)
				return nil, err
			}
			return con, err
		},
	}

	go ad.listenSubscription(ad.pubsubName)
	return &ad
}

type RedisAdapter struct {
	pubsubName string

	pool        *redis.Pool
	subscribers cmap.ConcurrentMap
	log         gopolling.Log
}

func (r *RedisAdapter) listenSubscription(name string) {
	subcon := redis.PubSubConn{Conn: r.pool.Get()}
	if err := subcon.Subscribe(name); err != nil {
		r.log.Errorf("fail to subscribe to %v, error: %v", name, err)
	}

	for {
		switch v := subcon.Receive().(type) {
		case redis.Message:
			var msg gopolling.Message
			if err := json.Unmarshal(v.Data, &msg); err != nil {
				r.log.Errorf("fail to unmarshal message, error: %v", err)
			}

			if val, ok := r.subscribers.Get(msg.Channel); ok {
				subMap := val.(*cmap.ConcurrentMap)
				subMap.IterCb(func(_ string, v interface{}) {
					ch := v.(chan gopolling.Message)
					ch <- msg
				})
			}
		case error:
			r.log.Errorf("error on subscription, error: %v", v)
			return
		case redis.Subscription:
			if v.Kind == "unsubscribe" {
				return
			}
		}
	}
}

func (r *RedisAdapter) Publish(_ string, msg gopolling.Message) error {
	con := r.pool.Get()

	st, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if _, err := con.Do("PUBLISH", r.pubsubName, st); err != nil {
		return err
	}
	return con.Close()
}

func (r *RedisAdapter) Subscribe(channel string) (gopolling.Subscription, error) {
	id := xid.New().String()
	ch := make(chan gopolling.Message)

	if val, ok := r.subscribers.Get(channel); ok {
		sub := val.(*cmap.ConcurrentMap)
		sub.Set(id, ch)
	} else {
		sub := cmap.New()
		sub.Set(id, ch)
		r.subscribers.Set(channel, &sub)
	}

	return gopolling.NewDefaultSubscription(channel, id, ch), nil
}

func (r *RedisAdapter) Unsubscribe(sub gopolling.Subscription) error {
	rsub := sub.(*gopolling.DefaultSubscription)
	if val, ok := r.subscribers.Get(rsub.Channel); ok {
		subMap := val.(*cmap.ConcurrentMap)
		subMap.Remove(rsub.ID)
	}

	return nil
}

func (r *RedisAdapter) Enqueue(channel string, ev gopolling.Event) {
	con := r.pool.Get()
	data, _ := json.Marshal(ev)
	if _, err := con.Do("RPUSH", channel, data); err != nil {
		r.log.Errorf("fail to perform RPUSH, error: %v", err)
	}
	if err := con.Close(); err != nil {
		r.log.Errorf("fail to close redis connection, error: %v", err)
	}
}

func (r *RedisAdapter) Dequeue(channel string) <-chan gopolling.Event {
	con := r.pool.Get()
	ch := make(chan gopolling.Event)

	go func() {
		for {
			if vals, err := redis.Values(con.Do("BLPOP", channel, 0)); err == nil {
				// first val is key, second is value
				val := vals[1].([]byte)
				var ev gopolling.Event
				if err := json.Unmarshal(val, &ev); err == nil {
					ch <- ev
				} else {
					r.log.Errorf("fail to  unmarshal ev, error: %v", err)
				}
			} else {
				r.log.Errorf("redis error: %v", err)
			}
		}
	}()

	return ch
}

func (r *RedisAdapter) SetLog(l gopolling.Log) {
	r.log = l
}

func (r *RedisAdapter) Find(key string) (gopolling.Message, bool) {
	var msg gopolling.Message

	con := r.pool.Get()
	b, err := redis.Bytes(con.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return msg, false
		}
		r.log.Errorf("redis fail to get key, error: %v", err)
		return msg, false
	}
	if err := con.Close(); err != nil {
		r.log.Errorf("fail to close redis connection, error: %v", err)
		return msg, false
	}

	if err := json.Unmarshal(b, &msg); err != nil {
		r.log.Errorf("fail to unmarshal, error: %v", err)
		return msg, false
	}

	return msg, true
}

func (r *RedisAdapter) Save(key string, msg gopolling.Message, t int) {
	con := r.pool.Get()

	st, err := json.Marshal(msg)
	if err != nil {
		r.log.Errorf("fail to marshal, error: %v", err)
		return
	}
	if _, err := con.Do("SETEX", key, t, st); err != nil {
		r.log.Errorf("redis fail to setex, error: %v", err)
		return
	}
	if err := con.Close(); err != nil {
		r.log.Errorf("fail to close redis connection, error: %v", err)
	}
}
