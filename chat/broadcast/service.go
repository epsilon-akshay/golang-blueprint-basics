package broadcast

import (
	"fmt"
	"github.com/gorilla/websocket"
)

type Room struct {
	ConnList []*websocket.Conn
	TrueConn map[*websocket.Conn]bool
	Users    map[*websocket.Conn]string
	Messages chan Message
}

type Message struct {
	by      string
	message string
}

func (room *Room) AddUser(con *websocket.Conn, userName string) {
	room.ConnList = append(room.ConnList, con)
	room.Users[con] = userName
	room.TrueConn[con] = true
	fmt.Println("adding new user", userName)
}

func (room *Room) RemoveUser(con *websocket.Conn) {
	room.TrueConn[con] = false
}

func (room *Room) RecieveAndForwardMessage() {
	for {
		msg := <-room.Messages
		for _, conn := range room.ConnList {
			connection := conn
			if room.TrueConn[connection] == true {
				fmt.Println(room.Users[connection])
				connection.WriteMessage(websocket.TextMessage, []byte(msg.message))
			}
		}
	}
}

func (room *Room) WriteMessage(conn *websocket.Conn) {
	for {
		_, message, _ := conn.ReadMessage()
		msg := Message{
			message: string(message),
			by: room.Users[conn],
		}
		room.Messages <- msg
	}
}
