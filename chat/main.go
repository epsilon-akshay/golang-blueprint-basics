package main

import (
	"github.com/epsilon-akshay/go-programming-blueprints/chat/broadcast"
	"github.com/gorilla/websocket"
	"net/http"
)

func roomHandler(room *broadcast.Room) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userName := r.URL.Query().Get("user_name")
		var upgrader = websocket.Upgrader{}
		ws, _ := upgrader.Upgrade(w, r, nil)

		room.AddUser(ws, userName)
		defer room.RemoveUser(ws)
		go room.RecieveAndForwardMessage()

		room.WriteMessage(ws)
	})
}

func main() {
	room := &broadcast.Room{
		TrueConn: make(map[*websocket.Conn] bool),
		ConnList: make([]*websocket.Conn, 1000),
		Users: make(map[*websocket.Conn] string),
		Messages: make(chan broadcast.Message),
	}
	http.HandleFunc("/", roomHandler(room))
	http.ListenAndServe(":8080", nil)
}
