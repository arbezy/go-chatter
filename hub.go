package main

import (
	"bytes"
	"html/template"
	"log"
	"sync"
)

type Hub struct {
	mu sync.RWMutex

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

// TODO: add a lock to hub (and acquire lock in this func)
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

			log.Printf("client registered &s", client.id)

			for _, msg := range h.messages {
				client.send <- getMessageTemplate(msg)
			}
		case client := <-h.unregister:
			// first check if exists
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				close(client.send)
				log.Printf("client unregistered %s", client.id)

				delete(h.clients, client)
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.mu.RLock()
			h.messages = append(h.messages, msg)

			for client := range h.clients {
				select {
				case client.send <- getMessageTemplate(msg):
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
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
