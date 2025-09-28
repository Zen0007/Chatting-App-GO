package router

import (
	"context"
	"log"
	"main/db"
	"main/upgrader"
	"main/utils"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SendContactRealTime(c *gin.Context) {
	u := upgrader.Upgrader()
	conn, err := u.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{utils.Err: err.Error()})
		return
	}

	defer conn.Close()

	var id User
	if err := conn.ReadJSON(&id); err != nil {
		log.Println(err.Error())
		return
	}

	out := make(chan any)
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range out {
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteJSON(msg); err != nil {
				log.Println("ws write error:", err)
				cancel()
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.D{
				{Key: "operationType", Value: "update"},
				{Key: "fullDocument.userId", Value: id.ID},
				{Key: "updateDescription.updatedFields.contact", Value: bson.D{{Key: "$exists", Value: true}}},
			}}},
			{{Key: "$project", Value: bson.D{
				{Key: "contact", Value: "$fullDocument.contact"},
			}}},
		}
		opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
		stream, err := db.Connect().Collection("user").Watch(ctx, pipeline, opts)
		if err != nil {
			log.Println("error when stream", err.Error())
			out <- gin.H{utils.Err: err.Error()}
			cancel()
			return
		}
		defer stream.Close(ctx)

		for stream.Next(ctx) {
			var event Friend

			if err := stream.Decode(&event); err != nil {
				log.Println("decode error", err.Error())
				continue
			}
			select {
			case out <- event.Contact:
			case <-ctx.Done():
				return
			}
		}

		if err := stream.Err(); err != nil {
			log.Println(err.Error())
			out <- gin.H{utils.Err: err.Error()}
			cancel()
		}

	}()

	// --- Wait for cancellation ---
	<-ctx.Done()

	// cleanup: close channel before waiting goroutines
	close(out)
	wg.Wait()
}
