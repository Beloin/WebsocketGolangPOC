package main

import "fmt"

type Hub struct {
	clients map[*Client]bool

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	SendMessage func(client *Client, message []byte)

	SendMessage2 chan []byte
}

func newHub() *Hub {
	fun := func(c *Client, message []byte) {
		fmt.Println("Got Message:")
		str := string(message)
		fmt.Println(str)
	}

	return &Hub{clients: make(map[*Client]bool), SendMessage: fun, SendMessage2: make(chan []byte)}
}

func (h *Hub) run() {
	println("Running HUB")
	for {
		select {
		case client := <-h.register:
			println("Registing client")
			h.clients[client] = true
		//case client := <-h.unregister:
		//	if _, ok := h.clients[client]; ok {
		//		delete(h.clients, client)
		//		close(client.send)
		//	}
		case message := <-h.SendMessage2:
			fmt.Printf(string(message))
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
