package router

import (
	"context"
	"fmt"
	"log"
	"main/db"
	"main/upgrader"
	"main/utils"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RealTimeChat(c *gin.Context) {
	u := upgrader.Upgrader()
	conn, err := u.Upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{utils.Err: err.Error()})
		fmt.Print(err.Error())
		return
	}

	defer conn.Close()

	out := make(chan any)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range out {
			if err := conn.WriteJSON(msg); err != nil {
				log.Println("error when write data to client", err.Error())
				cancel()
				break
			}
		}
	}()

	var firstMsg Message
	if err := conn.ReadJSON(&firstMsg); err != nil {
		log.Println(err.Error())
		cancel()
		close(out)
		wg.Wait()
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.D{
				{Key: "operationType", Value: "update"},
				{Key: "fullDocument.userId", Value: bson.D{
					{Key: "$in", Value: bson.A{firstMsg.Sender, firstMsg.Receiver}},
				}},
			}}},
			{{Key: "$project", Value: bson.D{
				{Key: "contactMessage", Value: "$updateDescription.updatedFields"},
			}}},
		}
		opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

		stream, err := db.Connect().Collection("user").Watch(ctx, pipeline, opts)
		if err != nil {
			log.Println("error when stream", err.Error())
			select {
			case out <- gin.H{utils.Err: err.Error()}:
			case <-ctx.Done():
			}
			return
		}
		defer stream.Close(ctx)

		for stream.Next(ctx) {
			var event messages

			if err := stream.Decode(&event); err != nil {
				log.Println("decode error", err.Error())
				continue
			}

			for _, v := range event.ContactMessage {
				select {
				case out <- gin.H{
					utils.Success: gin.H{
						"_id":      v.ID,
						"dateTime": v.DateTime,
						"receiver": v.Receiver,
						"sender":   v.Sender,
						"text":     v.Text,
					},
				}:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := stream.Err(); err != nil {
			log.Println(err.Error())
		}
	}()

	// --- Wait for cancellation ---
	<-ctx.Done()

	// cleanup: close channel before waiting goroutines
	close(out)
	wg.Wait()
}
