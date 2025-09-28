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
			var userID User
			if err := conn.ReadJSON(&userID); err != nil {
				log.Println("when readJson Get Contact", err.Error())
				cancel()
				break
			}

			filter := bson.M{
				"userId": userID.ID,
			}
			csr, err := db.Connect().Collection("user").Find(ctx, filter)
			if err != nil {
				log.Println(err.Error())
				cancel()
				break
			}

			for csr.Next(ctx) {
				var contact Friend
				if err := csr.Decode(&contact); err != nil {
					log.Println("Get contact", err.Error())
					break
				}
				select {
				case out <- contact.Contact:
				case <-ctx.Done():
					return
				}
			}

		}
	}()

	//
	<-ctx.Done()
	close(out)
	wg.Wait()
}
