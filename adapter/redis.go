package adapter

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"github.com/stanlry/gopolling"
	"time"
)

func safePushToChannel(ch chan gopolling.Message, message gopolling.Message) (result bool) {
	defer func() {
		if r := recover(); r != nil {
			result = false
		}
	}()
	ch <- message
	return true
}

func NewRedisAdapter(uri, password string) *RedisAdapter {
	p := RedisAdapter{
		log: &gopolling.NoOpLog{},
	}

	p.pool = redis.Pool{
		MaxIdle:     1,
		MaxActive:   100,
		IdleTimeout: 5 * time.Minute,
		Dial: func() (redis.Conn, error) {
			option := redis.DialPassword(password)
			con, err := redis.Dial("tcp", uri, option)
			if err != nil {
				p.log.Errorf("fail to connect to redis, error: %v", err)
				return nil, err
			}
			return con, err
		},
	}

	return &p
}

func newRedisSubscription(con redis.Conn, log gopolling.Log, roomID string) (gopolling.Subscription, error) {
	subcon := redis.PubSubConn{Conn: con}
	if err := subcon.Subscribe(roomID); err != nil {
		return nil, err
	}

	s := RedisSubscription{
		con: subcon,
		log: log,
		ch:  make(chan gopolling.Message),
	}

	go s.startListen()

	return &s, nil
}

type RedisSubscription struct {
	con redis.PubSubConn
	log gopolling.Log
	ch  chan gopolling.Message
}

func (r *RedisSubscription) startListen() {
loop:
	for {
		switch v := r.con.Receive().(type) {
		case redis.Message:
			var msg gopolling.Message
			if err := json.Unmarshal(v.Data, &msg); err != nil {
				msg.Error = err
			}

			if !safePushToChannel(r.ch, msg) {
				break loop
			}
		case error:
			var msg gopolling.Message
			msg.Error = v

			if !safePushToChannel(r.ch, msg) {
				break loop
			}
		case redis.Subscription:
			if v.Kind == "unsubscribe" {
				break loop
			}
		}
	}

	if err := r.con.Close(); err != nil {
		r.log.Errorf("fail to close redis connection, error: %v", err)
	}
}

func (r *RedisSubscription) Receive() <-chan gopolling.Message {
	return r.ch
}

func (r *RedisSubscription) Unsubscribe() error {
	close(r.ch)
	return r.con.Unsubscribe()
}

type RedisAdapter struct {
	pool redis.Pool
	log  gopolling.Log
}

func (r *RedisAdapter) Find(key string) (gopolling.Message, bool) {
	var msg gopolling.Message

	con := r.pool.Get()
	b, err := redis.Bytes(con.Do("GET", key))
	if err != nil {
		r.log.Errorf("redis fail to get key, error: %v", err)
		return msg, false
	}
	if err := con.Close(); err != nil {
		r.log.Errorf("fail to close redis connection, error: %v", err)
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
		r.log.Errorf("fail to marshal message, error: %v", err)
		return
	}
	if _, err := con.Do("SETEX", key, t, st); err != nil {
		r.log.Errorf("redis fail to set, error: %v", err)
		return
	}
	if err := con.Close(); err != nil {
		r.log.Errorf("fail to close redis connection, error: %v", err)
	}
}

func (r *RedisAdapter) SetLog(l gopolling.Log) {
	r.log = l
}

func (r *RedisAdapter) Publish(roomID string, msg gopolling.Message) error {
	data, _ := json.Marshal(msg)
	con := r.pool.Get()
	if _, err := con.Do("PUBLISH", roomID, data); err != nil {
		return err
	}
	return con.Close()
}

func (r *RedisAdapter) Subscribe(roomID string) (gopolling.Subscription, error) {
	return newRedisSubscription(r.pool.Get(), r.log, roomID)
}

func (r *RedisAdapter) Enqueue(roomID string, task gopolling.Event) {
	con := r.pool.Get()
	data, _ := json.Marshal(task)
	if _, err := con.Do("RPUSH", roomID, data); err != nil {
		r.log.Errorf("fail to perform RPUSH", "error", err)
	}
	if err := con.Close(); err != nil {
		r.log.Errorf("fail to close redis connection, error: %v", err)
	}
}

func (r *RedisAdapter) Dequeue(roomID string) <-chan gopolling.Event {
	con := r.pool.Get()
	ch := make(chan gopolling.Event)

	go func() {
		for {
			if vals, err := redis.Values(con.Do("BLPOP", roomID, 0)); err == nil {
				// first val is key, second is value
				val := vals[1].([]byte)
				var task gopolling.Event
				if err := json.Unmarshal(val, &task); err == nil {
					ch <- task
				} else {
					r.log.Errorf("fail to  unmarshal task, error: %v", err)
				}
			} else {
				r.log.Errorf("redis error: %v", err)
			}
		}
	}()

	return ch
}
