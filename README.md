GoPolling
==================
Simple tool for handling long polling request on server side.

### Install
```bash
go get github.com/stanlry/gopolling
```

### Usage
Simple http handler and notify
```go
package main

import (
    "github.com/stanlry/gopolling"
    "github.com/stanlry/gopolling/adapter"
    "github.com/uber/zap"
    "time"
    "http"
)

const roomID = "xxxxx"

var mgr = gopolling.NewGoPolling(gopolling.Option{ 
    // set the timeout for each request, default 120s   
    Timeout: 1 * time.Minute,  
    // message bus adapter, if your application has multiple stateless instances, use redis adapter
    // default use goroutine
    Adapter: adapter.NewRedisAdapter(":6379", "password"), 
    // logger interface, currently support zap and logrus, default will not log any error
    Logger: zap.New(), 
})

// http handler function
func handler(w http.ResponseWriter, r *http.Request) {
    // other logics ....

    // this will block until receive a notice or timeout
    val, err := mgr.WaitForNotice(r.Context(), roomID, &gopolling.DataOption{
        Data: "request data",
    })
    if err != nil {
        // handle errors
    }

    // response to client
}

// Notify polling client
func NotifyMessage() {
    mgr.Notify(roomID, gopolling.Message{
        Data: "notify client",
        Error: nil,
    })
}
```

To listen to event when request made and reply immediately if possible
```go
// subscribe listener
mgr.SubscribeListener(roomID, func(ev gopolling.Event, cb *gopolling.Callback){
   // handle request

   // reply
    cb.Reply("to client", nil)
}) 
```

If need to specific the identity of the client in room and reply to that particular client only
```go
const roomID = "common"

mgr.WaitForNotice(r.Context(), roomID, &gopolling.DataOption{
    Data: data,
    // declare client identify
    Selector: gopolling.S{"id": "123"}
})

mgr.Notify(roomID, gopolling.Message{
    Data: "response",
    // set the selector with same fields that client declared
    Selector: gopolling.S{"id": "123"}
})

// for subscriber, it will always reply to that particular client if selector was specified
mgr.SubscribeListener(roomID, func(ev gopolling.Event, cb *gopolling.Callback){
	cb.Reply("to client", nil)
}
```

### Examples
* [simple](/examples/simple)