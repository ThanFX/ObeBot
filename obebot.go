package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/robfig/cron"
)

const (
	SLACK_CONNECT_URL       string = "https://slack.com/api/rtm.start?token="
	GOOGLE_SEARCH_URL       string = "https://www.googleapis.com/customsearch/v1?"
	GOOGLE_SEARCH_ATTR      string = "&searchType=image&as_filetype=png&as_filetype=jpg&fields=items(link)"
	GOOGLE_SEARCH_MAX_PAGES int    = 90
)

type Keys struct {
	Slack   string `json: "slack"`
	Google  string `json: "google"`
	Cse     string `json: "cse"`
	Channel string `json: "channel"`
}

type Results struct {
	Items []struct {
		Link string `json: "link"`
	}
}

var (
	keys Keys
	m    Message
)

func getImage(q string, keys Keys, max int) string {
	var res Results
	randomStart := strconv.Itoa(rand.Intn(max) + 1)
	randomLink := rand.Intn(10)
	url := GOOGLE_SEARCH_URL + "key=" + keys.Google + "&cx=" + keys.Cse + "&q=" + q + GOOGLE_SEARCH_ATTR +
		"&start=" + randomStart + "&num=10"
	log.Println(url)
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
	if len(res.Items) < 10 {
		return "Умерь свою буйную фантазию!"
	}
	return res.Items[randomLink].Link
}

func postRandImage(ws *websocket.Conn, ch string) {
	//G0AM6NYU8 G3URW8HV2
	m.Type = "message"
	m.Channel = ch
	m.Text = getImage("обед", keys, GOOGLE_SEARCH_MAX_PAGES)
	postMessage(ws, m)
}

func main() {
	fmt.Println("Hello, I'm ObeBot!!")
	bs, err := ioutil.ReadFile("prop.json")
	if err != nil {
		log.Fatalf("Ошибка открытия файла параметров %s", err)
	}
	err = json.Unmarshal(bs, &keys)
	if err != nil {
		log.Fatalf("Ошибка получения параметров из JSON %s", err)
	}
	ws, id := slackConnect(keys.Slack)

	c := cron.New()
	c.AddFunc("0 0-30/5 11 * * MON-FRI", func() { postRandImage(ws, keys.Channel) })
	c.Start()

	rand.Seed(time.Now().UTC().UnixNano())
	for {
		m, err = getMessage(ws)
		if err != nil {
			log.Printf("Ошибка получения сообщения %s", err)
		}
		//log.Printf("Id: %d, Type: %s, Channel: %s, Text: %s", m.Id, m.Type, m.Channel, m.Text)
		// Смотрим только личные сообщения нашему Обеботу
		if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+id+">") {
			go func(m Message) {
				// Чем больше слов в запросе, тем меньший размер выборки (для повышения релевантности)
				length := 10 - len(strings.Fields(m.Text))
				if length < 1 {
					length = 1
				}
				//log.Println(length)
				//log.Println(math.Max(0, 5-length))
				// Преобразуем сообщение в поисковую строку, получим по запросу ссылку на картинку и запостим
				qs := strings.Join(strings.Fields(m.Text)[1:], "%20")
				link := getImage(qs, keys, length)
				m.Text = link
				postMessage(ws, m)
			}(m)
		}
	}
}