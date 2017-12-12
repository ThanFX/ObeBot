package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
)

type Photos struct {
	Category string
	Tags     []string
	Title    string
	Urls     []string
	PostUrl  string
}

var baseUrl string = "http://redbust.com/page/"
var startPage int = 1
var finishPage int = 492
var db *sql.DB

func savePost(post Photos) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Ошибка начала транзакции: %s", err)
	}
	res, err := tx.Exec("insert into posts(category, title, url) values(?, ?, ?)", post.Category, post.Title, post.PostUrl)
	if err != nil {
		log.Fatalf("Ошибка записи в таблицу posts: %s", err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatalf("Ошибка получения идентификатора добавленного поста: %s", err)
	}
	for i := 0; i < len(post.Tags); i++ {
		_, err := tx.Exec("insert into tags(post_id, tag) values(?, ?)", int(lastId), post.Tags[i])
		if err != nil {
			log.Fatalf("Ошибка записи в таблицу tags: %s", err)
		}
	}
	stmt, err := tx.Prepare("insert into photos(post_id, num, url) values(?, ?, ?)")
	if err != nil {
		log.Fatalf("Ошибка создания предварительной вставки: %s", err)
	}
	defer stmt.Close()
	for i := 0; i < len(post.Urls); i++ {
		_, err = stmt.Exec(lastId, i+1, post.Urls[i])
		if err != nil {
			log.Fatalf("Ошибка записи в таблицу photos: %s", err)
		}
	}
	tx.Commit()
	fmt.Printf("Записали пост №%d\n", lastId)
}

func getBlogUrl(pageUrl string) {
	fmt.Println("Старт парсинга страницы " + pageUrl)
	doc, err := goquery.NewDocument(pageUrl)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".post-row>article").Each(func(i int, blog *goquery.Selection) {
		getPhotos(blog.Find(".post-title>a").AttrOr("href", ""))
	})
}

func getPhotos(blogUrl string) {
	var photo Photos
	fmt.Println("Старт парсинга блога " + blogUrl)
	doc, err := goquery.NewDocument(blogUrl)
	if err != nil {
		log.Fatal(err)
	}

	photo.Category = doc.Find(".category>a").Text()
	photo.Title = doc.Find(".post-title").Text()
	photo.PostUrl = blogUrl
	doc.Find(".post-tags").First().Find("a").Each(func(i int, tag *goquery.Selection) {
		photo.Tags = append(photo.Tags, tag.Text())
	})
	doc.Find(".entry img").Each(func(i int, img *goquery.Selection) {
		photo.Urls = append(photo.Urls, img.AttrOr("src", ""))
	})
	savePost(photo)
}

func createDB() error {
	var err error
	err = os.Remove("redbust.db")
	if err != nil {
		log.Println(err)
	}
	db, err = sql.Open("sqlite3", "redbust.db")
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
		create table posts (id integer primary key, category text, title text, url text);
		create table photos (post_id integer not null, num integer, url text);
		create table tags (post_id integer not null, tag text);
		`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func main() {
	err := createDB()
	if err != nil {
		log.Println("БД создана!")
	}
	defer db.Close()
	for page := startPage; page <= finishPage; page++ {
		url := baseUrl + strconv.Itoa(page) + "/"
		getBlogUrl(url)
		time.Sleep(time.Second * 3)
	}
	log.Println("Парсинг закончен!")
}
