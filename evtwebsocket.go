package evtwebsocket

import (
	"errors"
	"time"

	"golang.org/x/net/websocket"
)

// Conn is the connection structure.
type Conn struct {
	OnMessage        func([]byte, *Conn)
	OnError          func(error)
	OnConnected      func(*Conn)
	MatchMsg         func([]byte, []byte) bool
	Reconnect        bool
	PingMsg          []byte
	PingIntervalSecs int
	ws               *websocket.Conn
	url              string
	subprotocol      string
	closed           bool
	msgQueue         []Msg
	pingTimer        time.Time
}

// Msg is the message structure.
type Msg struct {
	Body     []byte
	Callback func([]byte, *Conn)
}

// Dial sets up the connection with the remote
// host provided in the url parameter.
// Note that all the parameters of the structure
// must have been set before calling it.
func (c *Conn) Dial(url, subprotocol string) error {
	c.closed = true
	c.url = url
	c.subprotocol = subprotocol
	c.msgQueue = []Msg{}
	var err error
	c.ws, err = websocket.Dial(url, subprotocol, "http://localhost/")
	if err != nil {
		return err
	}
	c.closed = false
	if c.OnConnected != nil {
		go c.OnConnected(c)
	}

	go func() {
		defer c.close()

		for {
			var msg = make([]byte, 512)
			var n int
			if n, err = c.ws.Read(msg); err != nil {
				if c.OnError != nil {
					c.OnError(err)
				}
				return
			}
			c.onMsg(msg[:n])
		}
	}()

	c.setupPing()

	return nil
}

// Send sends a message through the connection.
func (c *Conn) Send(msg Msg) error {
	if c.closed {
		return errors.New("closed connection")
	}
	if _, err := c.ws.Write(msg.Body); err != nil {
		c.close()
		if c.OnError != nil {
			c.OnError(err)
		}
		return err
	}
	if c.PingIntervalSecs > 0 && c.PingMsg != nil {
		c.pingTimer = time.Now().Add(time.Second * time.Duration(c.PingIntervalSecs))
	}
	if msg.Callback != nil {
		c.msgQueue = append(c.msgQueue, msg)
	}

	return nil
}

// IsConnected tells wether the connection is
// opened or closed.
func (c *Conn) IsConnected() bool {
	return !c.closed
}

func (c *Conn) onMsg(msg []byte) {
	if c.MatchMsg != nil {
		for i, m := range c.msgQueue {
			if m.Callback != nil && c.MatchMsg(msg, m.Body) {
				go m.Callback(msg, c)
				// Delete this element from the queue
				c.msgQueue = append(c.msgQueue[:i], c.msgQueue[i+1:]...)
				break
			}
		}
	}
	// Fire OnMessage every time.
	if c.OnMessage != nil {
		go c.OnMessage(msg, c)
	}
}

func (c *Conn) close() {
	c.ws.Close()
	c.closed = true
	if c.Reconnect {
		for {
			if err := c.Dial(c.url, c.subprotocol); err == nil {
				break
			}
			time.Sleep(time.Second * 1)
		}
	}
}

func (c *Conn) setupPing() {
	if c.PingIntervalSecs > 0 && len(c.PingMsg) > 0 {
		c.pingTimer = time.Now().Add(time.Second * time.Duration(c.PingIntervalSecs))
		go func() {
			for {
				if !time.Now().After(c.pingTimer) {
					time.Sleep(time.Millisecond * 100)
					continue
				}
				if c.Send(Msg{c.PingMsg, nil}) != nil {
					return
				}
				c.pingTimer = time.Now().Add(time.Second * time.Duration(c.PingIntervalSecs))
			}
		}()
	}
}
