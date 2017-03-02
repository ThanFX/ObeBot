package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Question struct {
	Id        int
	Question  string
	Answer    string
	StartTime int64
	Attempt   int
}

type Result struct {
	AnswerCount int
	AnswerTime  int
}

var (
	quizLines  []string
	maxQ       int
	resultList map[string]Result
)

func initQuiz() bool {
	var fr *os.File
	//var v interface{}
	file, err := os.Open(FILE_QUIZ_NAME)
	if err != nil {
		log.Println("Отсутствует файл ", FILE_QUIZ_NAME, ", запуск викторины невозможен")
		return false
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	for sc.Scan() {
		quizLines = append(quizLines, sc.Text())
	}
	maxQ = len(quizLines)
	fmt.Printf("Успешно прочитали %d строк\n", maxQ)
	resultList = make(map[string]Result)

	if _, err = os.Stat(FILE_QUIZ_RESULT_NAME); !os.IsNotExist(err) {
		if fr, err = os.OpenFile(FILE_QUIZ_RESULT_NAME, os.O_RDONLY, 0777); err != nil {
			log.Printf("Ошибка открытия файла с результатами: %s", err)
		}
		e := json.NewDecoder(fr)
		if err = e.Decode(&resultList); err != nil {
			log.Printf("Ошибка парсинга файла с результатами: %s", err)
		}
	} else {
		log.Printf("Ошибка открытия файла с результатами: %s", err)
	}

	return true
}

func postQuiz(ws *websocket.Conn, m Message, isQuestion bool) bool {
	var (
		err    error
		prompt string
	)
	if !isQuestion {
		q, err = getNewQuestion()
		if err != nil {
			m.Text = "Ошибка! Ошибка!! Алярм!!!"
			postMessage(ws, m)
			return isQuestion
		}
		m.Text = q.Question
		postMessage(ws, m)
		isQuestion = true
	} else {
		q.Attempt++
		prompt, isQuestion = getQuestionHelp(q.Answer, q.Attempt)
		m.Text = "Повторяю вопрос: " + q.Question + prompt
		postMessage(ws, m)
	}
	return isQuestion
}

func getQuestionHelp(answer string, attepmt int) (string, bool) {
	if attepmt >= (len(answer) / 2) {
		return " Правильный ответ - " + answer, false
	} else {
		return " Подсказка - " + answer[:attepmt*2] + "...", true
	}
}

func getNewQuestion() (Question, error) {
	var (
		q Question
		d int64
	)
	random := rand.Intn(maxQ)
	n := 10
	d = 200
	// !!ToDo Протестировать и отрефакторить !!
	for {
		s := strings.Split(quizLines[random], "|")
		// Если вопрос ещё не задавался (нет даты задавания)
		if len(s) == 2 {
			q.Id = random
			q.Question = s[0]
			q.Answer = s[1]
			q.StartTime = time.Now().Unix()
			q.Attempt = 0
			break
		} else if len(s) == 3 {
			// Иначе считаем сколько времени уже прошло с момента последнего задавания
			lt, err := strconv.ParseInt(s[2], 10, 64)
			if err != nil {
				log.Printf("Ошибка парсинга времени задавания вопроса №%d", random+1)
				continue
			}
			dTime := (time.Now().Unix() - lt) * 3600 * 24
			// Если больше заданного порога - задаём
			if dTime > d {
				q.Id = random
				q.Question = s[0]
				q.Answer = s[1]
				q.StartTime = time.Now().Unix()
				q.Attempt = 0
				break
			} else {
				// Иначе делаем ещё одну попытку
				n--
			}
			// Если попытки закончились - уменьшаем порог на 10% и восстанавливаем попытки
			if n < 1 {
				n = 10
				d -= (d / 10)
			}
			// Если порог сильно уменьшился, выходим с ошибкой
			if d < 10 {
				return q, errors.New("Ответы на все вопросы были получены совсем недавно")
			}
		} else {
			log.Printf("Ошибка парсинга вопроса №%d \"%s\"", random+1, s[0])
		}
		random = rand.Intn(maxQ)
	}
	return q, nil
}

func checkAnswer(ws *websocket.Conn, m Message) bool {
	if strings.Contains(strings.ToLower(m.Text), q.Answer) {
		name, realName := getUserInfo(keys.Slack, m.User)
		postName := "Незнакомец"
		if realName == "" {
			postName = name
		} else {
			postName = realName
		}
		answerTime := time.Now().Unix() - q.StartTime
		m.Text = "Правильно, " + postName + ", это действительно так! Ответ: \"" + q.Answer + "\". Ответ дан за " + strconv.FormatInt(answerTime, 10) + "c"
		postMessage(ws, m)
		saveAnswerResult(m.User, answerTime)
		return false
	}
	return true
}

func saveAnswerResult(name string, answerTime int64) {
	var r Result
	var f *os.File
	if _, err := resultList[name]; !err {
		resultList[name] = Result{1, int(answerTime)}
	} else {
		r.AnswerCount = resultList[name].AnswerCount + 1
		r.AnswerTime = resultList[name].AnswerTime + int(answerTime)
		resultList[name] = r
	}
	// Если файла с результатами нету - создаём, иначе открываем. Записываем туда результаты
	if _, err := os.Stat(FILE_QUIZ_RESULT_NAME); os.IsNotExist(err) {
		f, err = os.Create(FILE_QUIZ_RESULT_NAME)
		if err != nil {
			log.Printf("Ошибка создания файла с результатми викторины: %s", err)
		}
		f.Close()
	}
	f, err := os.OpenFile(FILE_QUIZ_RESULT_NAME, os.O_WRONLY, 0444)
	if err != nil {
		log.Printf("Ошибка открытия файла с результатами викторины: %s", err)
	}
	defer f.Close()
	e := json.NewEncoder(f)
	err = e.Encode(resultList)
	if err != nil {
		log.Printf("Ошибка сохранения файла с результатами викторины: %s", err)
	}
}

func postQuizResult(ws *websocket.Conn, m Message) {
	//sortutil.DescByField(resultList, "AnswerCount")
	s := ""
	i := 1
	for key, value := range resultList {
		name, realName := getUserInfo(keys.Slack, key)
		postName := "Незнакомец"
		if realName == "" {
			postName = name
		} else {
			postName = realName
		}
		s += strconv.Itoa(i) + ". " + postName + ": дано ответов - " + strconv.Itoa(value.AnswerCount) + ", среднее время ответа - " +
			strconv.Itoa(value.AnswerTime/value.AnswerCount) + "с\n"
		i++
	}
	m.Text = s
	postMessage(ws, m)
}
