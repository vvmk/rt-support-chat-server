package main

import (
	r "github.com/dancannon/gorethink"
	"github.com/mitchellh/mapstructure"
)

const (
	ChannelStop = iota //sweet
	UserStop
	MessageStop
)

/* Handler signature :
 * type Handler func(*Client, interface{})
 */

/* from dancannon/gorethink/readme

   +Run - returns a cursor which can be used to view all rows returned.
   +RunWrite - returns a WriteResponse and should be used for queries such as Insert, Update, etc...
   +Exec - sends a query to the server and closes the connection immediately after reading the response from the database. If you do not wish to wait for the response then you can set the NoReply flag.*/

func addChannel(client *Client, data interface{}) {
	var channel Channel //load the data into this
	err := mapstructure.Decode(data, &channel)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}

	// this is a slow, blocking function, dispatch that shit to a goroutine
	// don't afraid, they're very light!
	go func() {
		insertErr := r.Table("channel").
			Insert(channel).
			Exec(client.session)
		if insertErr != nil {
			client.send <- Message{"error", err.Error()}
		}
	}()
}

/* subscribeChannel()
 *
 */
func subscribeChannel(client *Client, data interface{}) {
	stop := client.NewStopChannel(ChannelStop)
	result := make(chan r.ChangeResponse)

	cursor, err := r.Table("channel").
		Changes(r.ChangesOpts{IncludeInitial: true}).
		Run(client.session)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}

	go func() {
		var change r.ChangeResponse
		for cursor.Next(&change) {
			result <- change //send change to the result gochannel
		}
	}()

	go func() {
		for {
			select {
			case <-stop:
				cursor.Close()
				return
			case change := <-result:
				if change.NewValue != nil && change.OldValue == nil {
					client.send <- Message{"channel add", change.NewValue}
				}
			}
		}
	}()
}

func unsubscribeChannel(client *Client, data interface{}) {
	client.StopForKey(ChannelStop)
}

func unsubscribeUser(client *Client, data interface{}) {
	client.StopForKey(UserStop)
}

func addMessage(client *Client, data interface{}) {
	var chatMessage ChatMessage
	err := mapstructure.Decode(data, &chatMessage)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}

	go func() {
		insertErr := r.Table("message").
			Insert(chatMessage).
			Exec(client.session)
		if insertErr != nil {
			client.send <- Message{"error", insertErr.Error()}
		}
	}()
}

func unsubscribeMessage(client *Client, data interface{}) {
	client.StopForKey(MessageStop)
}
