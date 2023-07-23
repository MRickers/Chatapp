package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}
var clients = make(map[*websocket.Conn]bool)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	clients[conn] = true

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message: ", err)
			break
		}
		log.Println("read message: ", p)
		// Broadcast
		for client := range clients {
			log.Println("broadcasting message")
			err := client.WriteMessage(messageType, p)
			if err != nil {
				log.Println("Error writing message:", err)
				break
			}
		}
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../static/index.html")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ws", handleWebSocket)
	r.HandleFunc("/", homeHandler)

	fmt.Println("Chat Server started. Listening on port 8080...")
	log.Fatal(http.ListenAndServe(":5000", r))
}
