package main

import (
	"fmt"
	"log"
	"time"

	"github.com/rgamba/evtwebsocket"
)

func main() {
	c := evtwebsocket.Conn{

		// When connection is established
		OnConnected: func(w *evtwebsocket.Conn) {
			log.Println("Connected")
		},

		// When a message arrives
		OnMessage: func(msg []byte, w *evtwebsocket.Conn) {
			log.Printf("OnMessage: %s\n", msg)
		},

		// When the client disconnects for any reason
		OnError: func(err error) {
			log.Printf("** ERROR **\n%s\n", err.Error())
		},

		// This is used to match the request and response messagesP>termina
		MatchMsg: func(req, resp []byte) bool {
			return string(req) == string(resp)
		},

		// Auto reconnect on error
		Reconnect: true,

		// Set the ping interval (optional)
		PingIntervalSecs: 5,

		// Set the ping message (optional)
		PingMsg: []byte("PING"),
	}

	// Connect
	if err := c.Dial("ws://echo.websocket.org", ""); err != nil {
		log.Fatal(err)
	}

	for i := 1; i <= 100; i++ {

		// Create the message with a callback
		msg := evtwebsocket.Msg{
			Body: []byte(fmt.Sprintf("Hello %d", i)),
			Callback: func(resp []byte, w *evtwebsocket.Conn) {
				log.Printf("[%d] Callback: %s\n", i, resp)
			},
		}

		log.Printf("[%d] Sending message: %s\n", i, msg.Body)

		// Send the message to the server
		if err := c.Send(msg); err != nil {
			log.Println("Unable to send: ", err.Error())
		}

		// Take a break
		time.Sleep(time.Second * 2)
	}

}
