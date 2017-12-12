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

func postRandBoob2() string {
	var actress string = ""
	var res []BBRes
	randomNum := rand.Intn(BOOBS_MAX_SIZE)
	url := BOOBS_SEARCH_URL + strconv.Itoa(randomNum) + "/1/rank"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Ошибка запроса OBoobs: %s", err)
		return "Что-то ты не то ищещь, даже oboobs не хочет тебе отвечать"
	}
	if resp.StatusCode != 200 {
		return "А вот " + strconv.Itoa(resp.StatusCode) + " тебе!"
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка парсинга тела ответа от oboobs: %s", err)
	}
	defer resp.Body.Close()
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Printf("Ошибка парсинга ответа от oboobs: %s", err)
	}
	if res[0].Model != "" {
		actress = "Модель " + res[0].Model + ": "
	}
	return actress + "http://media.oboobs.ru/boobs/" + res[0].Preview[strings.LastIndex(res[0].Preview, "/")+1:len(res[0].Preview)]
}

func postRandButt2() string {
	var actress string = ""
	var res []BBRes
	randomNum := rand.Intn(BUTTS_MAX_SIZE)
	url := BUTTS_SEARCH_URL + strconv.Itoa(randomNum) + "/1/rank"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Ошибка запроса OButts: %s", err)
		return "Что-то ты не то ищещь, даже obutts не хочет тебе отвечать"
	}
	if resp.StatusCode != 200 {
		return "А вот " + strconv.Itoa(resp.StatusCode) + " тебе!"
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка парсинга тела ответа от obutts: %s", err)
	}
	defer resp.Body.Close()
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Printf("Ошибка парсинга ответа от obutts: %s", err)
	}
	if res[0].Model != "" {
		actress = "Модель " + res[0].Model + ": "
	}
	return actress + "http://media.obutts.ru/butts/" + res[0].Preview[strings.LastIndex(res[0].Preview, "/")+1:len(res[0].Preview)]
}

func postRandBoob(ws *websocket.Conn, m Message) {
	randomNum := rand.Intn(BOOBS_MAX_SIZE)
	var randomStr string
	log.Println(randomNum)
	switch {
	case randomNum < 10:
		randomStr = "0000" + strconv.Itoa(randomNum)
	case randomNum < 100:
		randomStr = "000" + strconv.Itoa(randomNum)
	case randomNum < 1000:
		randomStr = "00" + strconv.Itoa(randomNum)
	case randomNum < 10000:
		randomStr = "0" + strconv.Itoa(randomNum)
	default:
		randomStr = strconv.Itoa(randomNum)
	}
	link := "http://media.oboobs.ru/boobs/" + randomStr + ".jpg"
	m.Text = link
	postMessage(ws, m)
}

func postRandButt(ws *websocket.Conn, m Message) {
	randomNum := rand.Intn(BUTTS_MAX_SIZE)
	log.Println(randomNum)
	var randomStr string
	switch {
	case randomNum < 10:
		randomStr = "0000" + strconv.Itoa(randomNum)
	case randomNum < 100:
		randomStr = "000" + strconv.Itoa(randomNum)
	case randomNum < 1000:
		randomStr = "00" + strconv.Itoa(randomNum)
	case randomNum < 10000:
		randomStr = "0" + strconv.Itoa(randomNum)
	default:
		randomStr = strconv.Itoa(randomNum)
	}
	link := BUTTS_SEARCH_URL + randomStr + ".jpg"
	m.Text = link
	postMessage(ws, m)
}

func postBB(ws *websocket.Conn, m Message) {
	m.Text = postRandBoob2()
	postMessage(ws, m)
	m.Text = postRandButt2()
	postMessage(ws, m)
}
