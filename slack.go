package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
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

type User struct {
	Name     string `json:"name"`
	RealName string `json:"real_name"`
}

type responseUser struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
	User  User   `json: "user"`
}

type Channel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type responseChannel struct {
	Ok           bool   `json:"ok"`
	Error        string `json:"error"`
	ChannelsList []Channel
}

type Message struct {
	Id      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	User    string `json:"user"`
	Text    string `json:"text"`
	Ts      string `json:"ts"`
}

var counter uint64

func postMessage(ws *websocket.Conn, m Message) error {
	m.Id = atomic.AddUint64(&counter, 1)
	fmt.Println(m)
	return websocket.WriteJSON(ws, m)
}

func getMessage(ws *websocket.Conn) (m Message, err error) {
	_, s, err := ws.ReadMessage()
	err = json.Unmarshal(s, &m)
	return
}

func getUserInfo(token, user string) (string, string) {
	var respObj responseUser
	url := SLACK_GET_USER_INFO_URL + token + "&user=" + user
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Ошибка получения пользователя %s: %s", user, err)
	}
	if resp.StatusCode != 200 {
		log.Printf("Ошибка запроса %d при получении пользователя %s", resp.StatusCode, user)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка получения тела ответа при получении пользователя %s: %s", user, err)
	}
	resp.Body.Close()
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		log.Printf("Ошибка парсинга тела ответа от Slack при получении пользователя %s: %s", user, err)
	}
	if !respObj.Ok {
		log.Printf("Ошибка Slack при получении пользователя %s: %s", user, respObj.Error)
	}
	return respObj.User.Name, respObj.User.RealName
}

func getChannelsList(token string) []Channel {
	var respObj responseChannel
	url := SLACK_GET_CHANNELS_LIST_URL + token

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Ошибка получения списка каналов: %s", err)
	}
	if resp.StatusCode != 200 {
		log.Printf("Ошибка запроса %d при получении списка каналов: %s", resp.StatusCode, err)
	}
	log.Println(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка получения тела ответа при получении списка каналов: %s", err)
	}
	resp.Body.Close()
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		log.Printf("Ошибка парсинга тела ответа от Slack при получении списка каналов: %s", err)
	}
	if !respObj.Ok {
		log.Printf("Ошибка Slack при получении списка каналов: %s", respObj.Error)
	}
	//log.Println(respObj)
	return respObj.ChannelsList
}

func slackStart(token string) (wsurl, id string, err error) {
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

func slackConnect(token string) (*websocket.Conn, string) {
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
