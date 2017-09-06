package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"log"
	"net/http"
)

type Message struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type Channel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	//route requests to specified url to handler
	http.HandleFunc("/", handler)

	//block the app and listen for connections on specified
	http.ListenAndServe(":4000", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	for {
		var inMessage Message
		var outMessage Message
		if inErr := socket.ReadJSON(&inMessage); inErr != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("%#v\n", inMessage)
		switch inMessage.Name {
		case "channel add":
			addErr := addChannel(inMessage.Data)
			if addErr != nil {
				outMessage = Message{"error", addErr}
				if sockErr := socket.WriteJSON(outMessage); sockErr != nil {
					fmt.Println(err)
					break
				}
			}
		case "channel subscribe":
			// SLEPT HERE
		}
	}
}

func addChannel(data interface{}) error {
	var channel Channel
	err := mapstructure.Decode(data, &channel)
	if err != nil {
		return err
	}
	channel.Id = "1"
	fmt.Printf("%#v\n", channel)
	return nil
}
