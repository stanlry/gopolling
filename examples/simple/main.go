package main

import (
	"encoding/json"
	"github.com/stanlry/gopolling"
	"log"
	"net/http"
	"time"
)

var room = "test"

var mgr = gopolling.NewGoPolling(gopolling.Option{
	Timeout: 120 * time.Second,
})

func main() {
	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		data, err := mgr.WaitForNotice(r.Context(), room, &gopolling.DataOption{})
		if err != nil {
		} // handle error
		st, err := json.Marshal(data)
		if err != nil {
		} // handle error
		w.Write(st)
	})

	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		err := mgr.Notify(room, gopolling.Message{
			Data:  r.URL.Query().Get("data"),
			Error: nil,
		})
		if err != nil {
		} // handle error
	})

	log.Println("start serve on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
