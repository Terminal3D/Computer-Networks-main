package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
)

type URL struct {
	URL string `json:"url"`
}

func handleConnection(c *websocket.Conn) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		scanner.Scan()
		url := URL{scanner.Text()}
		data, _ := json.Marshal(url)
		err := c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Fatal("write:", err)
		}

		_, message, err := c.ReadMessage()
		if err != nil {
			log.Fatal("read:", err)
		}
		fmt.Println(string(message))
	}
}

func main() {
	c, resp, err := websocket.DefaultDialer.Dial("ws://localhost:8080", nil)

	if err != nil {
		log.Printf("handshake failed with status %d", resp.StatusCode)
		log.Fatal("dial:", err)
	}

	defer c.Close()
	handleConnection(c)
}
