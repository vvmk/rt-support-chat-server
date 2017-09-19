# rt-support-chat-server
A real-time chat app server written in Go. The front-end was created with React and can be found [here](https://github.com/vvmk/rt-support-chat).

I wrote this alongside a course by [James Moore](https://github.com/knowthen)

### planned features: 
+ edit/delete channels
+ private messages
+ authentication

### issues
+ possible data race on subscribeChannel and subscribeUser. I think this is handled but the error still appears when running with race flag
