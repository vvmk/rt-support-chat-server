package main

import (
	r "github.com/dancannon/gorethink"
	"github.com/mitchellh/mapstructure"
)

/* Handler signature :
 * type Handler func(*Client, interface{})
 */

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
