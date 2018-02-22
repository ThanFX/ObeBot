package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"

	"database/sql"

	"github.com/robfig/cron"
)

const (
	SLACK_CONNECT_URL           string = "https://slack.com/api/rtm.start?token="
	SLACK_GET_USER_INFO_URL     string = "https://slack.com/api/users.info?token="
	SLACK_GET_CHANNELS_LIST_URL string = "https://slack.com/api/channels.list?token="
	GOOGLE_SEARCH_URL           string = "https://www.googleapis.com/customsearch/v1?"
	GOOGLE_SEARCH_ATTR          string = "&searchType=image&as_filetype=png&as_filetype=jpg&fields=items(link)"
	GOOGLE_SEARCH_MAX_PAGES     int    = 90
	WIKI_SEARCH_URL             string = "https://ru.wikipedia.org/w/api.php?action=opensearch&prop=info&format=json&inprop=url"
	BOOBS_SEARCH_URL            string = "http://api.oboobs.ru/boobs/"
	BOOBS_MAX_SIZE              int    = 11240
	BUTTS_SEARCH_URL            string = "http://api.obutts.ru/butts/"
	BUTTS_MAX_SIZE              int    = 5230
	FILE_QUIZ_NAME              string = "quiz.txt"
	FILE_QUIZ_RESULT_NAME       string = "quiz_result.txt"
	BB_CHANNEL                  string = "G7XBF35TM"
)

type Keys struct {
	Slack      string `json: "slack"`
	Google     string `json: "google"`
	Cse        string `json: "cse"`
	Channel    string `json: "channel"`
	DialogFlow string `json: "dialogflow"`
}

type Results struct {
	Items []struct {
		Link string `json: "link"`
	}
}

type WikiRes struct {
	SearchText    string
	ResultStrings []string
	ResultDesc    []string
	ResultURL     []string
}

type BBRes struct {
	Model   string
	Preview string
	Id      int
	Rank    int
	Author  string
}

var (
	keys Keys
	m    Message
	q    Question
	db   *sql.DB
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

	db, err = sql.Open("sqlite3", "redbust.db")
	if err != nil {
		log.Printf("Ошибка открытия файла БД: %s", err)
	}
	defer db.Close()

	c := cron.New()
	c.AddFunc("0 0-30/5 6 * * MON-FRI", func() { postRandImage(ws, keys.Channel) })
	c.Start()

	rand.Seed(time.Now().UTC().UnixNano())
	for {
		m, err = getMessage(ws)
		if err != nil && err.Error() == "PANIC!" {
			log.Printf("Произошла паника, выдерживаем паузу и перезапускаемся")
			time.Sleep(time.Second * 30)
			ws, id = slackConnect(keys.Slack)
			m.Channel = BB_CHANNEL
			m.Text = "Паника отловлена и обезврежена, сиськи спасены!"
			postMessage(ws, m)
		}
		/*
			if err != nil {
				log.Printf("Ошибка получения сообщения %s", err)
			}
		*/
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
				if len(text) > 1 && text[0] == "<@"+id+">" {
					switch text[1] {
					// Если запрос на квиз - уходим туда
					case "-quiz":
						// Если викторина не запущена - нафиг
						if !isQuiz {
							m.Text = "Какая, нафиг, викторина? Работать, блеать!"
							postMessage(ws, m)
						} else {
							// Уходим постить вопрос (новый или повторять уже заданный)
							isQuestion = postQuiz(ws, m, isQuestion)
						}
					// Если запрос результатов квиза - уходим в результаты
					case "-result":
						// Если викторина не запущена - нафиг
						if !isQuiz {
							m.Text = "Какие, нафиг, результаты викторины? Работать, блеать!"
							postMessage(ws, m)
						} else {
							// Уходим постить результаты викторины
							postQuizResult(ws, m)
						}
					// Если запрос следующего вопроса - пропускаем текущий
					case "-next":
						if isQuestion {
							isQuestion = postQuiz(ws, m, false)
						}
					// Если запрос на Вики - уходим искать там
					case "-wiki":
						postWiki(ws, m, text[2:])
					// Запрос эротической картинки по тэгу
					case "-tag":
						if m.Channel == BB_CHANNEL {
							postRedImage(ws, m, text[2:])
						}
					// Поиск картинки по запросу
					case "-img":
						postImage(ws, m, text[1:])
					// Иначе просто диалог с ботом
					default:
						PostDialogMessage(ws, m, text[1:])
					}
				} else {
					// Если это канал b&b - парсим сообщение
					if m.Channel == BB_CHANNEL {
						if strings.HasPrefix(m.Text, "!") {
							postBB(ws, m)
						}
					} else if isQuiz && isQuestion {
						// Иначе смотрим, запущен ли квиз и загадан ли вопрос
						isQuestion = checkAnswer(ws, m)
					}
				}
			}(m)
		}
	}
}
