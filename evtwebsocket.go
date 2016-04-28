package evtwebsocket

import (
	"errors"
	"time"

	"golang.org/x/net/websocket"
)

// Conn is the connection structure.
type Conn struct {
	OnMessage        func([]byte)
	OnError          func(error)
	OnConnected      func(*websocket.Conn)
	MatchMsg         func([]byte, []byte) bool
	Reconnect        bool
	PingMsg          []byte
	PingIntervalSecs int
	ws               *websocket.Conn
	url              string
	closed           bool
	msgQueue         []Msg
	pingTimer        time.Time
}

// Msg is the message structure.
type Msg struct {
	Body     []byte
	Callback func([]byte)
}

// Dial sets up the connection with the remote
// host provided in the url parameter.
// Note that all the parameters of the structure
// must have been set before calling it.
func (c *Conn) Dial(url string) error {
	c.closed = true
	c.url = url
	c.msgQueue = []Msg{}
	var err error
	c.ws, err = websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		return err
	}
	c.closed = false
	go c.OnConnected(c.ws)

	go func() {
		defer c.close()

		for {
			var msg = make([]byte, 512)
			var n int
			if n, err = c.ws.Read(msg); err != nil {
				c.OnError(err)
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
		c.OnError(err)
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
	if c.MatchMsg == nil {
		return
	}
	for i, m := range c.msgQueue {
		if m.Callback != nil && c.MatchMsg(msg, m.Body) {
			go m.Callback(msg)
			// Delete this element from the queue
			c.msgQueue = append(c.msgQueue[:i], c.msgQueue[i+1:]...)
			return
		}
	}
	// If we didn't find a propper callback we
	// just fire the OnMessage global handler
	go c.OnMessage(msg)
}

func (c *Conn) close() {
	c.ws.Close()
	c.closed = true
	if c.Reconnect {
		for {
			if err := c.Dial(c.url); err == nil {
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
