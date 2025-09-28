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
		c.JSON(http.StatusBadRequest, gin.H{utils.Err: err.Error()})
		fmt.Print(err.Error())
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
			var msg Message
			if err := conn.ReadJSON(&msg); err != nil {
				log.Println("read error:", err)
				select {
				case out <- gin.H{utils.Err: err.Error()}:
				case <-ctx.Done():
				}
				cancel()
				break
			}
			if err := processMessage(ctx, msg); err != nil {
				select {
				case out <- gin.H{utils.Err: err.Error()}:
				case <-ctx.Done():
				}
			}
		}
	}()

	//Channel for all outgoing messages
	<-ctx.Done()
	close(out)
	wg.Wait()

}

func processMessage(ctx context.Context, msg Message) error {
	id := utils.GenerateIDText(10)

	dbConn := db.Connect().Collection("user")

	update := bson.M{
		"$push": bson.M{
			"contact.$.messages": bson.M{
				"id":       id,
				"sender":   msg.Sender,
				"receiver": msg.Receiver,
				"dateTime": msg.Date,
				"text":     msg.Text,
			},
		},
	}

	// use transactions for atomitcity
	session, err := db.Connect().Client().StartSession()

	if err != nil {
		return err
	}

	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessionCtx mongo.SessionContext) (interface{}, error) {
		if _, er := dbConn.UpdateOne(sessionCtx, bson.M{
			"userId":           msg.Sender,
			"contact.friendId": msg.Receiver,
		}, update); er != nil {
			return nil, er
		}
		if _, er := dbConn.UpdateOne(sessionCtx, bson.M{
			"userId":           msg.Receiver,
			"contact.friendId": msg.Sender,
		}, update); er != nil {
			return nil, er
		}
		return nil, nil
	})
	return err
}
