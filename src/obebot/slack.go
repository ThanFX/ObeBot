package main

import (
	"fmt"
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
	"github.com/gorilla/websocket"
	"sync/atomic"
)

type responseSelf struct {
	Id string `json:"id"`
}

type responseRtmStart struct {
	Ok    bool         `json:"ok"`
	Error string       `json:"error"`
	Url   string       `json:"url"`
	Self  responseSelf `json:"self"`
}

type Message struct {
	Id      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

var counter uint64

func postMessage(ws *websocket.Conn, m Message) error {
	m.Id = atomic.AddUint64(&counter, 1)
	return websocket.WriteJSON(ws, m)
}

func getMessage(ws *websocket.Conn) (m Message, err error) {
	err = websocket.ReadJSON(ws, &m)
	return
}

func slackStart(token string) (wsurl, id string, err error)  {
	var respObj responseRtmStart
	url := fmt.Sprint(SLACK_CONNECT_URL + token)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Ошибка соединения с Slack %s", err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("Ошибка запроса к Slack %d при соединении", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка получения тела ответа от Slack при соединении %s", err)
	}
	resp.Body.Close()
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		log.Fatalf("Ошибка парсинга тела ответа от Slack при соединении %s", err)
	}
	if !respObj.Ok {
		log.Fatalf("Ошибка Slack при соединении %s", respObj.Error)
	}
	wsurl = respObj.Url
	id = respObj.Self.Id
	return
}

func slackConnect(token string) (*websocket.Conn, string)  {
	wsurl, id, err := slackStart(token)
	if err != nil {
		log.Fatalf("Ошибка соединения со Slack %s", err)
	}

	ws, _, err := websocket.DefaultDialer.Dial(wsurl, nil)
	if err != nil {
		log.Fatalf("Ошибка открытия вебсокетного соединения со Slack %s", err)
	}
	return ws, id
}