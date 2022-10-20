package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
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
	hub     = newHub()
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
	go hub.run()
	http.HandleFunc("/chatbot", handler)
}

type Client struct {
	conn *websocket.Conn
	hub  *Hub

	send chan []byte
}

func handler(w http.ResponseWriter, req *http.Request) {
	conn, err := GetConnection(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{conn: conn, hub: hub}
	hub.register <- client

	fmt.Println(client)

	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		//c.hub.SendMessage2 <- message
		c.hub.SendMessage(c, message)
	}
}

func GetConnection(w http.ResponseWriter, req *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, req, nil)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
