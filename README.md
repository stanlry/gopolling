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
	"log"
	"net/http"
	"time"
)

var room = "test"

var mgr = gopolling.NewGoPolling(gopolling.Option{})

func main() {
	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		data, _ := mgr.WaitForNotice(r.Context(), room, &gopolling.DataOption{})
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
    // message bus adapter, if your application has multiple stateless instances, use redis adapter
    // default use goroutine
    Adapter: adapter.NewRedisAdapter(":6379", "password"), 
    // logger interface, currently support zap and logrus, default will not log any error
    Logger: zap.New(), 
})
```

#### Polling Client
client that wait for notice from listener or notifier
```go
func Polling(w http.ResponseWriter, r *http.Request) {
    // some logic ....

    // this will block until receive a notice or timeout
    val, err := mgr.WaitForNotice(r.Context(), roomID, &gopolling.DataOption{
        // send the data to listener, it will be discarded if no listener exist
        Data: "request data",
        // specify client identity in the room, this selector is essential a string map (map[string]string)
        // at the same time, notifier will have to specify the selector as well
        Selector: gopolling.S{
            "id": "xxx",
            "name": "xxx",
        }
    })

    // response to client
}
```

#### Direct Notify
Notify all the client that have been waiting in the room
```go
mgr.Notify(roomID, gopolling.Message{
    // data being sent to client
    Data: "notify client",
    // error
    Error: nil,
    // selector that specify the receiving client, if no client match the selector, message will be discarded
    Selector: gopolling.S{
        "id": "xxx",
    }
})
```
#### Event Listener
Listen to event when request made and reply immediately if possible
```go
// subscribe listener
mgr.SubscribeListener(roomID, func(ev gopolling.Event, cb *gopolling.Callback){
    // event data
    xx := ev.Data.(your_struct)
    // logic here

    // reply to client immediately if needed
    // Only the matching client will be notified if selector is declared
    resp := []string{"test", "test"}
    cb.Reply(resp, nil)
}) 
```

### Examples
* [simple](/examples/simple)