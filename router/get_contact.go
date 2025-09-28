package router

import (
	"context"
	"log"
	"main/db"
	"main/upgrader"
	"main/utils"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type Friend struct {
	Contact []Field `bson:"contact"`
}

type Field struct {
	Name string `bson:"friendName" json:"friendName"`
	ID   string `bson:"friendId" json:"id"`
}

type User struct {
	ID string `json:"userId"`
}

func GetContact(c *gin.Context) {
	u := upgrader.Upgrader()
	conn, err := u.Upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{utils.Err: err.Error()})
		log.Println("websocket upgrade failed:", err)
		return
	}

	defer conn.Close()

	out := make(chan any)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range out {
			if err := conn.WriteJSON(msg); err != nil {
				log.Println("error when write data to client", err.Error())
				cancel()
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			var userID User
			if err := conn.ReadJSON(&userID); err != nil {
				log.Println("read error GetContact:", err)
				cancel()
				return
			}

			filter := bson.M{
				"userId": userID.ID,
			}
			csr, err := db.Connect().Collection("user").Find(ctx, filter)
			if err != nil {
				log.Println("db find error:", err)
				out <- gin.H{utils.Err: err.Error()}
				cancel()
				return
			}

			defer csr.Close(ctx)

			for csr.Next(ctx) {
				var contact Friend
				if err := csr.Decode(&contact); err != nil {
					log.Println("decode error GetContact:", err)
					continue
				}
				select {
				case out <- contact.Contact:
				case <-ctx.Done():
					return
				}
			}

			if err := csr.Err(); err != nil {
				log.Println("cursor error", err)
				cancel()
				return
			}

		}
	}()

	//
	<-ctx.Done()
	close(out)
	wg.Wait()
}
