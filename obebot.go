package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/robfig/cron"
)

const (
	SLACK_CONNECT_URL           string = "https://slack.com/api/rtm.start?token="
	SLACK_GET_USER_INFO_URL     string = "https://slack.com/api/users.info?token="
	SLACK_GET_CHANNELS_LIST_URL string = "https://slack.com/api/channels.list?token="
	GOOGLE_SEARCH_URL           string = "https://www.googleapis.com/customsearch/v1?"
	GOOGLE_SEARCH_ATTR          string = "&searchType=image&as_filetype=png&as_filetype=jpg&fields=items(link)"
	GOOGLE_SEARCH_MAX_PAGES     int    = 90
	FILE_QUIZ_NAME              string = "quiz.txt"
	FILE_QUIZ_RESULT_NAME       string = "quiz_result.txt"
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
	q    Question
)

func main() {
	var (
		isQuestion = false
		isQuiz     = false
	)
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
	isQuiz = initQuiz()

	c := cron.New()
	c.AddFunc("0 0-30/5 12 * * MON-FRI", func() { postRandImage(ws, keys.Channel) })
	c.Start()

	rand.Seed(time.Now().UTC().UnixNano())
	for {
		m, err = getMessage(ws)
		if err != nil {
			log.Printf("Ошибка получения сообщения %s", err)
		}
		/*
			log.Printf("Id: %d, Type: %s, Channel: %s, User: %s, Text: %s, Time: %s",
				m.Id, m.Type, m.Channel, m.User, m.Text, m.Ts)
			if m.User != "" {
				log.Println(getUserInfo(keys.Slack, m.User))
			}
		*/
		// Парсим сообщение
		if m.Type == "message" {
			go func(m Message) {
				if m.Text == "" {
					return
				}
				text := strings.Fields(m.Text)
				// Если боту в личку - смотрим первое слово самого сообщения
				log.Println(text)
				if text[0] == "<@"+id+">" {
					switch text[1] {
					// Если запрос на квиз - уходим туда
					case "!quiz":
						// Если викторина не запущена - нафиг
						if !isQuiz {
							m.Text = "Какая, нафиг, викторина? Работать, блеать!"
							postMessage(ws, m)
						} else {
							// Уходим постить вопрос (новый или повторять уже заданный)
							isQuestion = postQuiz(ws, m, isQuestion)
						}
					// Если запрос результатов квиза - уходим в результаты
					case "!result":
						// Если викторина не запущена - нафиг
						if !isQuiz {
							m.Text = "Какие, нафиг, результаты викторины? Работать, блеать!"
							postMessage(ws, m)
						} else {
							// Уходим постить результаты викторины
							postQuizResult(ws, m)
						}
					// Если запрос следующего вопроса - пропускаем текущий
					case "!next":
						if isQuestion {
							isQuestion = postQuiz(ws, m, false)
						}
					// Если запрос на добавление вопроса
					case "!add":
						log.Println(getChannelsList(keys.Slack))
					// Иначе это просто запрос на картинки
					default:
						postImage(ws, m, text[1:])
					}
				} else {
					// Иначе смотрим, запущен ли квиз и загадан ли вопрос
					if isQuiz && isQuestion {
						isQuestion = checkAnswer(ws, m)
					}
				}
			}(m)
		}
	}
}
