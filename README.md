GoPolling
==================
[![Github Action](https://github.com/stanlry/gopolling/workflows/Test%20GoPolling/badge.svg)](https://github.com/stanlry/gopolling/workflows/Test%20GoPolling/badge.svg)
[![codecov](https://codecov.io/gh/stanlry/gopolling/branch/master/graph/badge.svg)](https://codecov.io/gh/stanlry/gopolling)
[![Go Report Card](https://goreportcard.com/badge/github.com/stanlry/gopolling)](https://goreportcard.com/report/github.com/stanlry/gopolling)

Simple tool for handling long polling request on server side.

### Install
```bash
go get github.com/stanlry/gopolling
```

### Example
```go
package main

import (
    "encoding/json"
    "github.com/stanlry/gopolling"
    "log"
    "net/http"
)

var channel = "test"

var mgr = gopolling.New(gopolling.DefaultOption)

func main() {
    http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
        resp, _ := mgr.WaitForNotice(r.Context(), channel, nil)
        st, _ := json.Marshal(resp)
        w.Write(st)
    })
    
    http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
        data := r.URL.Query().Get("data")
        mgr.Notify(channel, data, nil, nil)
    })
        
    log.Println("start serve on :80")
    log.Fatal(http.ListenAndServe(":80", nil))
}
```
wait for message
```bash
curl -s localhsot/message
```
notify clients
```bash
curl -s localhost/notify?data=[your message here]
```

### Usage
#### Create Polling Manager
```go
var mgr = gopolling.New(gopolling.Option{ 
    // message retention time, default is 60s
    Retention: 60,

    // set the timeout for each request, default 120s   
    Timeout: 1 * time.Minute,  

    // message bus, default use goroutine, you can choose redis as messaging bus
    Bus: adapter.NewRedisAdapter(":6379", "password"), 

    // message buffer, default use memory, you can choose redis as buffer
    Buffer: adapter.NewRedisAdapter(":6379", "password"), 

    // logger interface, currently support zap and logrus, default will not log any error
    Logger: zap.New(), 
})
```

#### Wait For Notice
wait for notice from listener or notifier
```go
// this function will block until receive a notice or timeout
resp, err := mgr.WaitForNotice(
    // request context
    r.Context(), 
    // channel
    channel, 
    // send the data to listener, it will be discarded if no listener exist
    "data",
})
```
only wait for notice with matched selector
```go
resp, err := mgr.WaitForSelectedNotice(
    r.Context(),
    channel,
    "data",
    // specify identity in the channel, this selector is essential a string map
    gopolling.S{
        "id": "xxx",
        "name": "xxx",
    }
)
```

#### Direct Notify
Notify everyone that have been waiting in the channel
```go
mgr.Notify(
    // channel
    channel,
    // data being sent
    "data to notify client",
    // error
    nil,
    // selector that specify the receiving side, if no one match the selector, message will be discarded
    gopolling.S{
        "id": "xxx",
    },
)
```
#### Event Listener
Listen to event when request was made and reply immediately. The reply message will only notify the one who
make the request
```go
// subscribe listener
mgr.SubscribeListener(channel, func(ev gopolling.Event, cb *gopolling.Callback){
    // event data
    xx := ev.Data.(your_struct)

    // reply to immediately, you can skip this part if no reply is needed
    data := "hi there"
    cb.Reply(resp, nil)
}) 
```
