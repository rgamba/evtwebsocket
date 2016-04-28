# Event Websocket

evtwebsocket provides an extremely easy way of dealing with websocket connections as a client in an event oriented manner:

```go
conn = evtwebsocket.Conn{
    // Fires when the connection is established
    OnConnect: func(ws *ws.Conn) {
        fmt.Println("Connected!")   
    },
    // Fires when a new message arrives from the server
    OnMessage: func(msg []byte) {
        fmt.Printf("New message: %s\n", msg)   
    },
    // Fires when an error occurs and connection is closed
    OnError: func(err error) {
        fmt.Printf("Error: %s\n", err.Error())
        os.Exit(1)   
    },
}
```
It also provides an extremely easy way to match request and response messages to work with callbacks like:
```go
msg := evtwebsocket.Msg{
    Body: []byte("Message body"),
    Callback: func(resp []byte) {
        // This function executes when the server responds
        fmt.Printf("Got response: %s\n", resp)  
    }, 
}
conn.Send(msg)
```
To be able to match the request and response, you must provide a matching function (this can also be done in the conn creation along with OnConnect, OnMessage, etc.)
```go
conn.MatchMsg = func(req, resp []byte) {
    // This one assumes response messages will always echo requests
    return req == resp
}
```
Note that when a callback is set in the request message, if the response arrives, the OnMessage event will be overriden by the Callback event, and therefore OnMessage wont get fired.
OnMessage will only get fired when the message from the server doesn't satisfy the matching condition.