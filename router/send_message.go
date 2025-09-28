package router

import (
	"context"
	"fmt"
	"log"
	"main/db"
	"main/upgrader"
	"main/utils"
	"sync"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Text     string `json:"text"`
	Date     string `json:"date"`
}

type messages struct {
	ID struct {
		Data string `bson:"_data"`
	} `bson:"_id"`
	ContactMessage map[string]struct {
		ID       string `bson:"id"`
		DateTime string `bson:"dateTime"`
		Receiver string `bson:"receiver"`
		Sender   string `bson:"sender"`
		Text     string `bson:"text"`
	} `bson:"contactMessage"`
}

func SendMessage(c *gin.Context) {
	u := upgrader.Upgrader()
	conn, err := u.Upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		conn.WriteJSON(gin.H{utils.Err: err.Error()})
		fmt.Print(err.Error())
		return
	}

	out := make(chan any)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer conn.Close()
		defer wg.Done()
		for msg := range out {
			if err := conn.WriteJSON(msg); err != nil {
				log.Println("error when write data to client", err.Error())
				cancel()
				break
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			var messages Message
			if err := conn.ReadJSON(&messages); err != nil {
				select {
				case out <- gin.H{utils.Err: err.Error()}:
				case <-ctx.Done():
				}
				log.Println(err.Error())
				cancel()
				break
			}
			processMessage(messages, out, ctx)
		}
	}()

	//Channel for all outgoing messages
	<-ctx.Done()
	wg.Wait()
	close(out)

}

func processMessage(messages Message, out chan<- any, ctx context.Context) {
	id := utils.GenerateIDText(10)
	filterSend := bson.M{
		"userId":           messages.Sender,
		"contact.friendId": messages.Receiver,
	}
	updateSend := bson.M{
		"$push": bson.M{
			"contact.$.messages": bson.M{
				"id":       id,
				"sender":   messages.Sender,
				"receiver": messages.Receiver,
				"dateTime": messages.Date,
				"text":     messages.Text,
			},
		},
	}
	_, err := db.Connect().Collection("user").UpdateOne(ctx, filterSend, updateSend)
	if err != nil {
		select {
		case out <- gin.H{utils.Err: err.Error()}:
		case <-ctx.Done():
		}
		log.Println(err.Error())
		return
	}

	filterRec := bson.M{
		"userId":           messages.Receiver,
		"contact.friendId": messages.Sender,
	}
	updateRec := bson.M{
		"$push": bson.M{
			"contact.$.messages": bson.M{
				"id":       id,
				"sender":   messages.Sender,
				"receiver": messages.Receiver,
				"dateTime": messages.Date,
				"text":     messages.Text,
			},
		},
	}
	_, err = db.Connect().Collection("user").UpdateOne(ctx, filterRec, updateRec)
	if err != nil {
		select {
		case out <- gin.H{utils.Err: err.Error()}:
		case <-ctx.Done():
		}
		log.Println(err.Error())
		return
	}

}
