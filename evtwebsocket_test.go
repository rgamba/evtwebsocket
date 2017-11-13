package evtwebsocket

import (
	"testing"
	"time"
	"golang.org/x/net/websocket"
	"crypto/tls"
)

func TestConn_Dial(t *testing.T) {
	type args struct {
		url         string
		subprotocol string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"ws-normal",
			args{
				"ws://echo.websocket.org",
				"",
			},
			false,
		},
		{
			"ws-tls",
			args{
				"wss://echo.websocket.org",
				"",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{}
			if err := c.Dial(tt.args.url, tt.args.subprotocol); (err != nil) != tt.wantErr {
				t.Errorf("Conn.Dial() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConn_DialConfig(t *testing.T) {
	type args struct {
		url         string
		subprotocol string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"ws-normal",
			args{
				"ws://echo.websocket.org",
				"",
			},
			false,
		},
		{
			"ws-tls",
			args{
				"wss://echo.websocket.org",
				"",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{}
			config, _ := websocket.NewConfig(tt.args.url, "http://localhost")
			config.TlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
			config.Protocol = []string{tt.args.subprotocol}
			if err := c.DialConfig(config); (err != nil) != tt.wantErr {
				t.Errorf("Conn.Dial() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConn_Send(t *testing.T) {
	type fields struct {
		OnMessage   func([]byte, *Conn)
		OnError     func(error)
		OnConnected func(*Conn)
		MatchMsg    func([]byte, []byte) bool
	}
	type args struct {
		url string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"regular-send",
			fields{
				OnConnected: func(con *Conn) {
					m := Msg{
						Body: []byte("Hello"),
						Callback: func(msg []byte, con *Conn) {
							if string(msg) != "Hello" {
								t.Errorf("Callback() expected = 'Hello', got = '%s'", msg)
							}
						},
					}
					if err := con.Send(m); err != nil {
						t.Errorf("Conn.Send() error = %v", err)
					}
				},
				OnMessage: func(msg []byte, con *Conn) {
					if string(msg) != "Hello" {
						t.Errorf("OnMessage() expected = 'Hello', got = '%s'", msg)
					}
				},
				MatchMsg: func(req, resp []byte) bool {
					return string(req) == string(resp)
				},
				OnError: func(err error) {
					t.Errorf("Error: %v", err)
				},
			},
			args{
				"ws://echo.websocket.org",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				OnMessage:   tt.fields.OnMessage,
				OnConnected: tt.fields.OnConnected,
				MatchMsg:    tt.fields.MatchMsg,
			}
			err := c.Dial(tt.args.url, "")
			if err != nil {
				t.Errorf("Conn.Dial() error = %v", err)
			}
			// Wait for response
			time.Sleep(time.Second * 2)
		})
	}
}
