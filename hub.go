package main

import (
	"bytes"
	"html/template"
	"log"
)

type Hub struct {
	clients  map[*Client]bool
	messages []*Message

	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// not concurrent
// TODO: add a lock to hub (and acquire lock in this func)
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

			log.Printf("client registered &s", client.id)

			for _, msg := range h.messages {
				client.send <- getMessageTemplate(msg)
			}
		case client := <-h.unregister:
			// first check if exists
			if _, ok := h.clients[client]; ok {
				close(client.send)
				log.Printf("client unregistered %s", client.id)

				delete(h.clients, client)
			}
		case msg := <-h.broadcast:
			h.messages = append(h.messages, msg)

			for client := range h.clients {
				select {
				case client.send <- getMessageTemplate(msg):
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func getMessageTemplate(msg *Message) []byte {
	templ, err := template.ParseFiles("templates/message.html")
	if err != nil {
		log.Fatal("Template parsing: %s", err)
	}

	var renderedMessage bytes.Buffer
	err = templ.Execute(&renderedMessage, msg)
	if err != nil {
		log.Fatal("template parsing: %s", err)
	}

	return renderedMessage.Bytes()
}
