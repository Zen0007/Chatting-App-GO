package router

import (
	"context"
	"fmt"
	"log"
	"main/db"

	"main/upgrader"

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

func SendMessage(c *gin.Context) {
	u := upgrader.Upgrader()
	conn, err := u.Upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		conn.WriteJSON(gin.H{"err": err.Error()})
		fmt.Print(err.Error())
		return
	}

	go handlerSendMessage(conn)
	go func() {
		defer conn.Close()

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
			conn.WriteJSON(gin.H{"error stream": err.Error()})
			return
		}
		defer stream.Close(context.Background())

		for stream.Next(context.Background()) {
			var event map[string]any

			if err := stream.Decode(&event); err != nil {
				log.Println("decode error", err.Error())
				continue
			}
			if err := conn.WriteJSON(event); err != nil {
				log.Println("error when write data to client", err.Error())
			}

		}

		if err := stream.Err(); err != nil {
			log.Fatal(err.Error())
		}

	}()
	// select{}
}

func handlerSendMessage(conn *websocket.Conn) {
	defer conn.Close()

	for {
		var messages Message
		if err := conn.ReadJSON(&messages); err != nil {
			conn.WriteJSON(gin.H{"err": err.Error()})
			break
		}

		filterSend := bson.M{
			"userID":           messages.Sender,
			"contact.friendId": messages.Receiver,
		}
		updateSend := bson.M{
			"$push": bson.M{
				"contact.$.messages": bson.M{
					"sender":   messages.Sender,
					"receiver": messages.Receiver,
					"dateTime": messages.Date,
					"text":     messages.Text,
				},
			},
		}
		_, err := db.UpdateOne("user", filterSend, updateSend)
		if err != nil {
			conn.WriteJSON(gin.H{"err update sender": err.Error()})
			break
		}

		filterRec := bson.M{
			"userID":           messages.Receiver,
			"contact.friendId": messages.Sender,
		}
		updateRec := bson.M{
			"$push": bson.M{
				"contact.$.messages": bson.M{
					"sender":   messages.Sender,
					"receiver": messages.Receiver,
					"dateTime": messages.Date,
					"text":     messages.Text,
				},
			},
		}
		_, err = db.UpdateOne("user", filterRec, updateRec)
		if err != nil {
			conn.WriteJSON(gin.H{"err receiver": err.Error()})
			break
		}

		err = conn.WriteJSON(gin.H{
			"success": fmt.Sprintf("send text to: %s", messages.Receiver),
		},
		)
		if err != nil {
			conn.WriteJSON(gin.H{"err": err.Error()})
			break

		}

	}

}
