package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
)

func main() {
	name := "varsha"
	c, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:8080?user_name=%s", name), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	go func() {
		for {
			_,msg,err := c.ReadMessage()
			if err != nil {
				fmt.Print("error reading message")
			}
			fmt.Println(string(msg))
		}
	}()
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		err := c.WriteMessage(1,[]byte(text))
		if err != nil {
			log.Println("write:", err)
			return
		}
	}
}

