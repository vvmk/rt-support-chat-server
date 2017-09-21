package main

import (
	r "github.com/dancannon/gorethink"
	"github.com/mitchellh/mapstructure"
	"time"
	"fmt"
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

func editChannel(client *Client, data interface{}) {
	var channel Channel
	err := mapstructure.Decode(data, &channel)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}
	//fmt.Println("editChannel: ", channel.Id)
	go func() {
		_, updateErr := r.Table("channel").
			Get(channel.Id).
			Update(channel).
			RunWrite(client.session)
		if updateErr != nil {
			client.send <- Message{"error", updateErr.Error()}
			return
		}
	}()
}

func deleteChannel(client *Client, data interface{}) {
	var channel Channel
	err := mapstructure.Decode(data, &channel)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}
	go func() {
		deleteErr := r.Table("channel").
			Get(channel.Id).
			Delete().
			Exec(client.session)
		if deleteErr != nil {
			client.send <- Message{"error", deleteErr.Error()}
			return
		}
	}()
}
/* subscribeChannel()
 *
 */
func subscribeChannel(client *Client, data interface{}) {
	go func() {
		stop := client.NewStopChannel(ChannelStop)

		cursor, err := r.Table("channel").
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}
		changeFeedHelper(cursor, "channel", client.send, stop)
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
	client.userName = user.Name
	go func() {
		_, insertErr := r.Table("user").
			Get(client.id).
			Update(user).
			RunWrite(client.session)
		if insertErr != nil {
			client.send <- Message{"error", insertErr.Error()}
			return
		}
	}()
}

func subscribeUser(client *Client, data interface{}) {
	go func() {
		stop := client.NewStopChannel(UserStop)

		cursor, err := r.Table("user").
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)

		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}

		changeFeedHelper(cursor, "user", client.send, stop)
	}()
}

func unsubscribeUser(client *Client, data interface{}) {
	client.StopForKey(UserStop)
}

func addChatMessage(client *Client, data interface{}) {
	var chatMessage ChatMessage
	err := mapstructure.Decode(data, &chatMessage)
	if err != nil {
		client.send <- Message{"error", err.Error()}
		return
	}

	go func() {
		chatMessage.CreatedAt = time.Now()
		chatMessage.Author = client.userName

		insertErr := r.Table("message").
			Insert(chatMessage).
			Exec(client.session)
		if insertErr != nil {
			client.send <- Message{"error", insertErr.Error()}
		}
	}()
}

func subscribeChatMessage(client *Client, data interface{}) {
	go func() {
		eventData := data.(map[string]interface{})
		val, ok := eventData["channelId"]
		if !ok {
			return
		}
		channelId, ok := val.(string)
		if !ok {
			return
		}
		
		stop := client.NewStopChannel(MessageStop)

		// TODO: this will work but its not using indexes, needs to be 
		// 			optimized if used outside of demo context
		cursor, err := r.Table("message").
			OrderBy(r.OrderByOpts{Index: r.Asc("createdAt")}).
			Filter(r.Row.Field("channelId").Eq(channelId)).
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)
		if err != nil {
			client.send <- Message{"error", err.Error()}
			return
		}

		changeFeedHelper(cursor, "message", client.send, stop)
	}()
}

func unsubscribeChatMessage(client *Client, data interface{}) {
	client.StopForKey(MessageStop)
}

func changeFeedHelper(cursor *r.Cursor, changeEventName string,
send chan<- Message, stop <-chan bool) {
	change := make(chan r.ChangeResponse)
	cursor.Listen(change) //instead of change.next in a goroutine
	for {
		eventName := ""
		var data interface{}
		select {
		case <-stop:
			cursor.Close()
			return
		case val := <-change:
			if val.NewValue != nil && val.OldValue == nil {
				eventName = changeEventName + " add"
				data = val.NewValue
			} else if val.NewValue == nil && val.OldValue != nil {
				eventName = changeEventName + " remove"
				data = val.OldValue
			} else if val.NewValue != nil && val.OldValue != nil {
				eventName = changeEventName + " edit"
				data = val.NewValue
			}
			send <- Message{eventName, data}
		}
	}
}
