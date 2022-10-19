package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// Probably set this buffers size bigger than they are.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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

func GetConnection(w http.ResponseWriter, req *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, req, nil)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
