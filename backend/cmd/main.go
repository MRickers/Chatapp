package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}
var clients = make(map[*websocket.Conn]bool)
var userNames = make(map[*websocket.Conn]string)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	// Prompt user for username and store it
	conn.WriteMessage(websocket.TextMessage, []byte("Enter your username:"))
	_, username, err := conn.ReadMessage()
	if err != nil {
		log.Println("Error reading username:", err)
		return
	}

	userNames[conn] = string(username)
	clients[conn] = true

	updateUsersList()

	// Handle chat messages
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// Broadcast the received message to all connected clients
		for clientConn := range clients {
			err := clientConn.WriteMessage(messageType, p)
			if err != nil {
				log.Println("Error writing message:", err)
				break
			}
		}
	}

	delete(clients, conn)
	delete(userNames, conn)
	updateUsersList()
}

func updateUsersList() {
	userList := make([]string, 0, len(userNames))
	for _, username := range userNames {
		userList = append(userList, username)
	}

	userJSON, err := json.Marshal(userList)
	if err != nil {
		log.Println("Error marshaling user list:", err)
		return
	}

	for clientConn := range clients {
		clientConn.WriteMessage(websocket.TextMessage, userJSON)
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
	log.Fatal(http.ListenAndServe(":8080", r))
}
