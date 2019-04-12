package main

import (
	"log"
	"strings"

	"github.com/gorilla/websocket"

	"math/rand"
)

func getRandImage(postId int) (string, string) {
	max, err := db.Query("select count(*) from photos where post_id = ?", postId)
	if err != nil {
		log.Fatalf("Ошибка поиска количества фотографий по id в БД: %s", err)
	}
	defer max.Close()

	rows, err := db.Query("select url from photos where post_id = ?", postId)
	if err != nil {
		log.Fatalf("Ошибка поиска фотографий по id в БД: %s", err)
	}
	defer rows.Close()

	tags, err := db.Query("select tag from tags where post_id = ?", postId)
	if err != nil {
		log.Fatalf("Ошибка поиска фотографий по id в БД: %s", err)
	}
	defer tags.Close()

	var maxCount, randPhotoNum int
	for max.Next() {
		err = max.Scan(&maxCount)
		if err != nil {
			log.Fatalf("Ошибка парсинга количества найденных фотографий: %s", err)
		}
		randPhotoNum = rand.Intn(maxCount)
	}
	i := 0
	var url, s string
	for rows.Next() {
		err = rows.Scan(&url)
		if err != nil {
			log.Fatal("ошибка парсинга url фотографий: ", err)
		}
		if i == randPhotoNum {
			break
		}
		i++
	}

	var tags_string []string
	for tags.Next() {
		err = tags.Scan(&s)
		if err != nil {
			log.Fatal("ошибка парсинга tag фотографий: ", err)
		}
		tags_string = append(tags_string, s)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Выбрали фото №%d", randPhotoNum)
	return url, strings.Join(tags_string, ", ")
}

func getImageUrlByTag(tag string) string {
	log.Println("Ищем по запросу " + tag)
	max, err := db.Query("select count(post_id) from tags where tag like ?", strings.ToLower("%"+tag+"%"))
	if err != nil {
		log.Fatalf("Ошибка поиска количества постов по тегу в БД: %s", err)
	}

	defer max.Close()

	rows, err := db.Query("select post_id from tags where tag like ?", strings.ToLower("%"+tag+"%"))
	if err != nil {
		log.Fatalf("Ошибка поиска id постов по тегу в БД: %s", err)
	}

	defer rows.Close()

	var maxCount, randPostNum int
	for max.Next() {
		err = max.Scan(&maxCount)
		if err != nil {
			log.Fatalf("Ошибка парсинга количества найденных постов: %s", err)
		}
		if maxCount < 1 {
			return "Андрюха, уймись, нет такого тэга!"
		}
		randPostNum = rand.Intn(maxCount)
	}
	i := 0
	var id int
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			log.Fatal("ошибка парсинга id поста: ", err)
		}
		if i == randPostNum {
			break
		}
		i++
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Выбрали пост №%d", id)
	url, tags := getRandImage(id)
	return "Фото с тэгами " + tags + ": " + url
}

func postRedImage(ws *websocket.Conn, m Message, text []string) {
	// Преобразуем сообщение в поисковую строку, получим по запросу ссылку на картинку и запостим
	qs := strings.Join(text, " ")
	link := getImageUrlByTag(qs)
	m.Text = link
	m.Type = "message"
	postMessage(ws, m)

	//	if text[0] == "redhead" {
	//		m.Text = "<@U5SSW8KLM> :point_up:"
	//		postMessage(ws, m)
	//	}
}
