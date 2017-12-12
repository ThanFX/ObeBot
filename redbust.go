package main

import (
	"strings"
	"github.com/gorilla/websocket"
)

func postRedImage(ws *websocket.Conn, m Message, text []string) {
	// Преобразуем сообщение в поисковую строку, получим по запросу ссылку на картинку и запостим
	qs := strings.Join(text, "%20")
	link := getWiki(qs)
	m.Text = link
	m.Type = "message"
	postMessage(ws, m)
}
