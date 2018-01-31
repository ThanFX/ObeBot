package main

import (
	"log"
	"strings"

	"github.com/gorilla/websocket"
	. "github.com/mlabouardy/dialogflow-go-client"
	apiai "github.com/mlabouardy/dialogflow-go-client/models"
)

func getDialogAnswer(input string) string {
	err, dfc := NewDialogFlowClient(apiai.Options{
		AccessToken: keys.DialogFlow,
	})
	if err != nil {
		log.Fatal(err)
	}

	query := apiai.Query{
		Query: input,
	}
	//log.Printf("Запрос к диалогботу: %s", query)
	resp, err := dfc.QueryFindRequest(query)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%s", resp.Result.Fulfillment.Messages)
	if resp.Result.Fulfillment.Speech == "" {
		return "Моя твоя не понимай, выражайся яснее"
	} else {
		return resp.Result.Fulfillment.Speech
	}
}

func PostDialogMessage(ws *websocket.Conn, m Message, query []string) {
	m.Text = getDialogAnswer(strings.Join(query, " "))
	m.Type = "message"
	postMessage(ws, m)
}
