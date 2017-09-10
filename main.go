package main

import (
	r "github.com/dancannon/gorethink"
	"log"
	"net/http"
	"time"
)

type Channel struct {
	Id   string `json:"id" gorethink:"id,omitempty"`
	Name string `json:"name" gorethink:"name"`
}

type User struct {
	Id   string `json:"id" gorethink:"id,omitempty"`
	Name string `json:"name" gorethink:"name"`
}

type ChatMessage struct {
	Id        string    `json:"id" gorethink:"id,omitempty"`
	Author    string    `json:"author" gorethink:"author"`
	Body      string    `json:"body" gorethink:"body"`
	ChannelId string    `json:"channelId" gorethink:"channelId"`
	CreatedAt time.Time `json:"createdAt" gorethink:"createdAt"`
}

func main() {
	session, err := r.Connect(r.ConnectOpts{
		Address:  "localhost:28015",
		Database: "rtsupport",
	})
	if err != nil {
		log.Panic(err.Error())
	}

	router := NewRouter(session)

	router.Handle("channel add", addChannel)
	router.Handle("channel subscribe", subscribeChannel)
	router.Handle("channel unsubscribe", unsubscribeChannel)

	// router.Handle("user edit", editUser)
	// router.Handle("user subscribe", subscribeUser)
	router.Handle("user unsubscribe", unsubscribeUser)

	router.Handle("message add", addMessage)
	router.Handle("message subscribe", subscribeMessage)
	router.Handle("messagse unsubscribe", unsubscribeMessage)

	http.Handle("/", router)
	http.ListenAndServe(":4000", nil)
}
