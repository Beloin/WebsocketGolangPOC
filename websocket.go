package main

import (
	"bytes"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
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
}

func Server() {
	http.HandleFunc("/chatbot", handler)
}

type Client struct {
	conn *websocket.Conn
}

func handler(w http.ResponseWriter, req *http.Request) {
	conn, err := GetConnection(w, req)
	if err != nil {
		log.Println(err)
		return
	}

	client := Client{conn: conn}
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
		c.hub.broadcast <- message
	}
}

func GetConnection(w http.ResponseWriter, req *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, req, nil)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
