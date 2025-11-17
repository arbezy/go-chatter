package main

type Message struct {
	ClientId string      `json:"clientID"`
	Text     interface{} `json:"text"`
}
