GoPolling
==================
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
    "github.com/stanlry/gopolling/adapter"
    "log"
    "net/http"
)

var room = "test"

var mgr = gopolling.NewGoPolling(gopolling.Option{
    Bus: adapter.NewGoroutineAdapter(),
})

func main() {
    http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
        data, _ := mgr.WaitForNotice(r.Context(), room, nil)
        st, _ := json.Marshal(data)
        w.Write(st)
    })
    
    http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
        mgr.Notify(room, gopolling.Message{
            Data:  r.URL.Query().Get("data"), 
            Error: nil,
        })
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
var mgr = gopolling.NewGoPolling(gopolling.Option{ 
    // set the timeout for each request, default 120s   
    Timeout: 1 * time.Minute,  

    // message bus
    Bus: adapter.NewGoroutine(),
    // you can use redis as message bus
    // Bus: adapter.NewRedisAdapter(":6379", "password"), 

    // logger interface, currently support zap and logrus, default will not log any error
    Logger: zap.New(), 
})
```

#### Polling Client
client that wait for notice from listener or notifier
```go
// this function will block until receive a notice or timeout
val, err := mgr.WaitForNotice(
    // request context
    r.Context(), 
    // room is the classification, clients in the same room will all be notified if no selector is specified
    room, 
    // send the data to listener, it will be discarded if no listener exist
    "data",
})
```
client that only wait for notice with matched selector
```go
val err := mgr.WaitForSelectedNotice(
    r.Context(),
    room,
    "data",
    // specify client identity in the room, this selector is essential a string map
    gopolling.S{
        "id": "xxx",
        "name": "xxx",
    }
)
```

#### Direct Notify
Notify all the client that have been waiting in the room
```go
mgr.Notify(roomID, gopolling.Message{
    // data being sent to client
    Data: "data to notify client",
    // error
    Error: nil,
    // selector that specify the receiving client, if no client match the selector, message will be discarded
    Selector: gopolling.S{
        "id": "xxx",
    }
})
```
#### Event Listener
Listen to event when request was made and reply immediately. The reply message will only notify the client who
make the request
```go
// subscribe listener
mgr.SubscribeListener(roomID, func(ev gopolling.Event, cb *gopolling.Callback){
    // event data
    xx := ev.Data.(your_struct)

    // reply to client immediately, you can skip this part if no reply is needed
    resp := []string{"I'm", "data"}
    cb.Reply(resp, nil)
}) 
```
