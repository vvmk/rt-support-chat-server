package main

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/gorilla/websocket"
	"net/http"
)

type Handler func(*Client, interface{})

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Router struct {
	rules   map[string]Handler
	session *r.Session
}

/* maps must be initialized before they're used */
/* no constructors in go, this is the way to go */
func NewRouter(session *r.Session) *Router {
	return &Router{
		rules:   make(map[string]Handler),
		session: session,
	}
}

/* Handle(msgName string, handler Handler)
 * add a message handling rule to the rules map
 */
func (r *Router) Handle(msgName string, handler Handler) {
	r.rules[msgName] = handler
}

/* FindHandler(msgName string)
 * provide a way for the client to lookup and call handler funcs
 */
func (r *Router) FindHandler(msgName string) (Handler, bool) {
	handler, found := r.rules[msgName]
	return handler, found
}

func (e *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	client := NewClient(socket, e.FindHandler, e.session)
	defer client.Close() //covers our ass if something goes awry
	go client.Write()
	client.Read()
}
