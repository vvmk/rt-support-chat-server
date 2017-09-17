package main

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/gorilla/websocket"
	"log"
)

type FindHandler func(string) (Handler, bool)

type Message struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type Client struct {
	send         chan Message // a gochannel that passes Messages
	socket       *websocket.Conn
	findHandler  FindHandler
	session      *r.Session
	stopChannels map[int]chan bool
	id			 string
	userName	 string
}

/*
 * provide a way for the app to obtain a gochannel on which it
 * can listen for stop signals (through the stop chan
 */
func (c *Client) NewStopChannel(stopKey int) chan bool {
	c.StopForKey(stopKey) // prevent possible goroutine leak

	stop := make(chan bool)
	c.stopChannels[stopKey] = stop
	return stop
}

/* StopForKey(key int)
 * 		stop the change feed and exit the subscribe goroutines
 */
func (c *Client) StopForKey(key int) {
	if ch, found := c.stopChannels[key]; found {
		ch <- true
		delete(c.stopChannels, key)
	}
}

/* handle messages sent to client */
func (client *Client) Read() {
	var message Message
	for {
		if err := client.socket.ReadJSON(&message); err != nil {
			fmt.Println("client.Read Error: ", err)
			break
		}
		// router should expose a function that can look up a handler
		// then call that function here.
		if handler, found := client.findHandler(message.Name); found {
			handler(client, message.Data)
		}
	}
	client.socket.Close()
}

func (client *Client) Write() {
	for msg := range client.send {
		fmt.Printf("%#v\n", msg)
		if err := client.socket.WriteJSON(msg); err != nil {
			fmt.Println("client.Write Error: ", err)
			break
		}
	}
	client.socket.Close()
}

func (c *Client) Close() {
	for _, ch := range c.stopChannels {
		ch <- true
	}
	close(c.send)

	//probably a better place to do this
	r.Table("user").Get(c.id).Delete().Exec(c.session)
}

func NewClient(socket *websocket.Conn, findHandler FindHandler,
	session *r.Session) *Client {
	var user User
	user.Name = "anonymous"
	res, err := r.Table("user").Insert(user).RunWrite(session)
	if err != nil {
		log.Println(err.Error())
	}
	var id string
	if len(res.GeneratedKeys) > 0 {
		id = res.GeneratedKeys[0]
	}
	return &Client{
		send:         make(chan Message),
		socket:       socket,
		findHandler:  findHandler,
		session:      session,
		stopChannels: make(map[int]chan bool),
		id:			  id,
		userName:	  user.Name,
	}
}
