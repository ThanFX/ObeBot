package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

func getWiki(s string) string {
	var res []string
	url := WIKI_SEARCH_URL + "&search=" + s
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Ошибка запроса Wiki: %s", err)
		return "Что-то ты не то ищещь, даже Wiki не хочет тебе отвечать"
	}
	if resp.StatusCode != 200 {
		return "А вот " + strconv.Itoa(resp.StatusCode) + " тебе!"
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка парсинга тела ответа от Wiki: %s", err)
	}
	defer resp.Body.Close()
	bodyStr := string(body)
	answers := strings.Split(bodyStr[1:len(bodyStr)-1], "],")
	log.Println(answers[2])
	err = json.Unmarshal([]byte(answers[2]), &res)
	log.Println(res)
	if err != nil {
		log.Printf("Ошибка парсинга ответа от Wiki: %s", err)
	}
	if len(res) == 0 {
		return "Ты запрашиваешь настолько неведомую дичь, что даже Вика не в курсе"
	}
	randomLink := rand.Intn(len(res))
	return res[randomLink]
}

func postWiki(ws *websocket.Conn, m Message, text []string) {
	// Преобразуем сообщение в поисковую строку, получим по запросу ссылку на картинку и запостим
	qs := strings.Join(text, "%20")
	link := getWiki(qs)
	m.Text = link
	m.Type = "message"
	postMessage(ws, m)
}
