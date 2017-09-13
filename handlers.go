package main

import (
	r "github.com/dancannon/gorethink"
	"github.com/mitchellh/mapstructure"
	"time"
)

const (
	ChannelStop = iota //sweet
	UserStop
	MessageStop
)

func addChannel(client *Client, data interface{}) {
	var channel Channel //load the data into this
	err := mapstructure.Decode(data, &channel)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}

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

func editUser(client *Client, data interface{}) {
	var user User
	err := mapstructure.Decode(data, &user)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}

	go func() {
		response, insertErr := r.Table("user").
			Insert(user, r.InsertOpts{Conflict: "replace"}).
			RunWrite(client.session)
		if insertErr != nil {
			client.send <- Message{"error", insertErr.Error()}
			return
		}
	}()
}

func subscribeUser(client *Client, data interface{}) {
	stop := client.NewStopChannel(UserStop)
	result := make(chan r.ChangeResponse)

	cursor, err := r.Table("user").
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
					client.send <- Message{"user add", change.NewValue}
				} else if change.NewValue != nil && change.OldValue != nil {
					client.send <- Message{"user edit", change.NewValue}
				} else if change.NewValue == nil && change.OldValue != nil {
					client.send <- Message{"user remove", change.OldValue}
				}
			}
		}
	}()
}

func unsubscribeUser(client *Client, data interface{}) {
	client.StopForKey(UserStop)
	// TODO: remove client activeUser from database
}

func addMessage(client *Client, data interface{}) {
	var chatMessage ChatMessage
	err := mapstructure.Decode(data, &chatMessage)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}

	chatMessage.CreatedAt = time.Now()

	go func() {
		insertErr := r.Table("message").
			Insert(chatMessage).
			Exec(client.session)
		if insertErr != nil {
			client.send <- Message{"error", insertErr.Error()}
		}
	}()
}

func subscribeMessage(client *Client, data interface{}) {
	stop := client.NewStopChannel(MessageStop)
	result := make(chan r.ChangeResponse)

	var activeChannel ActiveChannel
	mapStructErr := mapstructure.Decode(data, &activeChannel)
	if mapStructErr != nil {
		client.send <- Message{"error", mapStructErr.Error()}
		return
	}

	// TODO: these need to be ordered by "createdAt". Rethink
	//initially appears not to have a straightforward way to do this.
	// try again later...
	cursor, err := r.Table("message").
		GetAllByIndex("channelId", activeChannel.ChannelId).
		//Between(activeChannel.ChannelId, activeChannel.ChannelId, r.BetweenOpts{Index: "compound", RightBound: "open", LeftBound: "open"}).
		//OrderBy(r.OrderByOpts{Index: "compound"}).
		Changes(r.ChangesOpts{IncludeInitial: true}).
		Run(client.session)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}

	go func() {
		var change r.ChangeResponse
		for cursor.Next(&change) {
			result <- change
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
					client.send <- Message{"message add", change.NewValue}
				}
			}
		}
	}()
}

func unsubscribeMessage(client *Client, data interface{}) {
	client.StopForKey(MessageStop)
}
