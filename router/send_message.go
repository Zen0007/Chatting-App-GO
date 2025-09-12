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
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
		conn.WriteJSON(gin.H{"err": err.Error()})
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
		handlerSendMessage(conn, ctx, out)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		watchMessage(ctx, out)

	}()
	//Channel for all outgoing messages
	<-ctx.Done()
	wg.Wait()
	close(out)
}

func watchMessage(ctx context.Context, out chan<- any) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "operationType", Value: "update"},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "contactMessage", Value: "$updateDescription.updatedFields"},
		}}},
	}

	stream, err := db.Connect().Collection("user").Watch(context.Background(), pipeline)
	if err != nil {
		log.Println("error when stream", err.Error())
		select {
		case out <- gin.H{"error stream": err.Error()}:
		case <-ctx.Done():
		}
		return
	}
	defer stream.Close(context.Background())

	for stream.Next(context.Background()) {
		var event messages

		if err := stream.Decode(&event); err != nil {
			log.Println("decode error", err.Error())
			continue
		}

		for _, v := range event.ContactMessage {
			select {
			case out <- gin.H{
				"_id":      v.ID,
				"dateTime": v.DateTime,
				"receiver": v.Receiver,
				"sender":   v.Sender,
				"text":     v.Text,
			}:
			case <-ctx.Done():
			}
		}
	}

	if err := stream.Err(); err != nil {
		log.Fatal(err.Error())
	}

}

func handlerSendMessage(conn *websocket.Conn, ctx context.Context, out chan<- any) {
	for {
		var messages Message
		if err := conn.ReadJSON(&messages); err != nil {
			select {
			case out <- gin.H{"err": err.Error()}:
			case <-ctx.Done():
			}
			break
		}

		id := utils.GenerateIdText(10)
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
		_, err := db.UpdateOne("user", filterSend, updateSend)
		if err != nil {
			select {
			case out <- gin.H{"err update sender": err.Error()}:
			case <-ctx.Done():
			}
			break
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
		_, err = db.UpdateOne("user", filterRec, updateRec)
		if err != nil {
			select {
			case out <- gin.H{"err receiver": err.Error()}:
			case <-ctx.Done():
			}
			break
		}

		select {
		case out <- gin.H{
			"success": fmt.Sprintf("send text to: %s", messages.Receiver),
		}:
		case <-ctx.Done():
		}

	}

}
