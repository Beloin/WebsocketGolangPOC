package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Probably set this buffers size bigger than they are.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// SEE:
// https://github.com/gorilla/websocket
// https://pkg.go.dev/github.com/gorilla/websocket
// https://github.com/gorilla/websocket/blob/af47554f343b4675b30172ac301638d350db34a5/examples/chat/client.go#L82
// https://github.com/gorilla/websocket/blob/af47554f343b4675b30172ac301638d350db34a5/examples/chat/hub.go#L9
// EXAMPLE: https://javascript.info/websocket#:~:text=To%20open%20a%20websocket%20connection,It's%20like%20HTTPS%20for%20websockets.
func main() {
	fmt.Println("Hello World")
	Server()
}

func Server() {
	hub := newHub()
	go hub.run()

	http.HandleFunc("/chatbot", func(w http.ResponseWriter, req *http.Request) {
		println("Started")
		handler(hub, w, req)
	})

	_ = http.ListenAndServe(":9091", nil)
}

type Client struct {
	conn *websocket.Conn
	hub  *Hub

	send chan []byte
}

func handler(hub *Hub, w http.ResponseWriter, req *http.Request) {
	conn, err := GetConnection(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}
	println("Connected")

	client := &Client{conn: conn, hub: hub, send: make(chan []byte)}

	go func() { client.hub.register <- client }()

	fmt.Println("Registered")

	go client.readPump()
	go client.writePump()
}

func GetConnection(w http.ResponseWriter, req *http.Request) (*websocket.Conn, error) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, req, nil)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) readPump() {
	defer func() {
		fmt.Println("Closed Client")
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		panic(err)
	}
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		println("Incoming Message:")
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
