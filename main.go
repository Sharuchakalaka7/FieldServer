package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)




var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var data [2][]byte
var idx int




func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Temperature: %s C\n", data[0])
	fmt.Fprintf(w, "Humidity:    %s %%\n", data[1])
}


func reader(conn *websocket.Conn) {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		if (len(p) > 0) {
			data[idx] = p
			idx = (idx+1) % len(data)
		}

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "WebSocket Endpoint")
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil { log.Println(err) }

	log.Println("Client Successfully Connected...")
	reader(ws)
}


func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	idx = 0
	fmt.Println("Go WebSockets")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
