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

func getImage(q string, keys Keys, max int) string {
	var res Results
	randomStart := strconv.Itoa(rand.Intn(max) + 1)
	url := GOOGLE_SEARCH_URL + "key=" + keys.Google + "&cx=" + keys.Cse + "&q=" + q + GOOGLE_SEARCH_ATTR +
		"&start=" + randomStart + "&num=10"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Ошибка запроса CSE: %s", err)
		return "Что-то ты не то ищещь, даже гугл не хочет тебе отвечать"
	}
	if resp.StatusCode != 200 {
		return "А вот " + strconv.Itoa(resp.StatusCode) + " тебе!"
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка парсинга тела ответа от CSE: %s", err)
	}
	defer resp.Body.Close()
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Printf("Ошибка парсинга ответа от CSE: %s", err)
	}
	if len(res.Items) < 3 {
		return "Умерь свою буйную фантазию!"
	}
	randomLink := rand.Intn(len(res.Items))
	return res.Items[randomLink].Link
}

func postRandImage(ws *websocket.Conn, ch string) {
	//G0AM6NYU8 G3URW8HV2
	m.Type = "message"
	m.Channel = ch
	m.Text = getImage("обед", keys, GOOGLE_SEARCH_MAX_PAGES)
	postMessage(ws, m)
}

func postImage(ws *websocket.Conn, m Message, text []string) {
	// Чем больше слов в запросе, тем меньший размер выборки (для повышения релевантности)
	length := 10 - len(text)
	if length < 1 {
		length = 1
	}
	// Преобразуем сообщение в поисковую строку, получим по запросу ссылку на картинку и запостим
	qs := strings.Join(text, "%20")
	link := getImage(qs, keys, length)
	m.Text = link
	m.Type = "message"
	postMessage(ws, m)
}
