package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type URL struct {
	URL string `json:"url"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		var urlReceived URL
		err = json.Unmarshal(p, &urlReceived)
		if err != nil {
			errorMsg := make([]byte, len(err.Error()))
			copy(errorMsg, err.Error())
			log.Println(err)
			conn.WriteMessage(messageType, errorMsg)
			continue
		}

		log.Println(urlReceived.URL)
		u, err := url.ParseRequestURI(urlReceived.URL)
		if err != nil {
			errorMsg := make([]byte, len(err.Error()))
			log.Println(err)
			copy(errorMsg, err.Error())
			conn.WriteMessage(messageType, errorMsg)
			continue
		}
		log.Println("sending request to", u)
		if response, err := http.Get(urlReceived.URL); err != nil {
			log.Println("request to", u, " failed", "error", err)
			continue
		} else {
			defer response.Body.Close()
			status := response.StatusCode
			log.Println("got response from", u, "status", status)
			if status == http.StatusOK {
				if doc, err := ioutil.ReadAll(response.Body); err != nil {
					log.Println("invalid HTML from", u, "error", err)
					continue
				} else {
					conn.WriteMessage(messageType, doc)
					log.Println("HTML result sent")
				}
			}
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
