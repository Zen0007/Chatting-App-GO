package router

import (
	"context"
	"fmt"
	"log"
	"main/db"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func handlerIncomingMessage(conn *websocket.Conn) {
	defer conn.Close()
	msg := <-update
	pipeline := mongo.Pipeline{
		{
			{
				Key: "$match", Value: bson.D{
					{
						Key: "operationType", Value: "update",
					},
					{
						Key: "fullDocument.username", Value: bson.D{
							{
								Key: "$in", Value: bson.A{msg.Sender, msg.Receiver},
							},
						},
					},
					{
						Key:   "updateDescription.updatedFields.contact.message",
						Value: bson.D{{Key: "$exists", Value: true}},
					},
				},
			},
		},
		{
			{
				Key: "$project", Value: bson.D{
					{
						Key: "fullDocument.contact.message", Value: 1,
					},
				},
			},
		},
	}

	stream, err := db.Collection(pipeline)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer stream.Close(context.Background())

	for stream.Next(context.Background()) {
		var event map[string]any
		if err := stream.Decode(&event); err != nil {
			log.Println("decode error", err.Error())
			continue
		}
		if err := conn.WriteJSON(gin.H{"data": event}); err != nil {
			log.Println(err.Error())
			continue
		}
	}

	if err := stream.Err(); err != nil {
		log.Fatal(err.Error())
	}
	defer close(update)
}
