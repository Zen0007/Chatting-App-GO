// Package router
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
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Contacts struct {
	UserID    string `json:"userId"`
	ContactID string `json:"contactId"`
}

type Ids struct {
	ID primitive.ObjectID `bson:"_id"`
}

type Friends struct {
	Name   string `bson:"username"`
	UserID string `bson:"userId"`
}

func AddContact(c *gin.Context) {
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
				log.Println("error when write", err.Error())
				cancel()
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		handlerAddContact(conn, ctx, out, cancel)
	}()

	// Wait until connection is closed
	<-ctx.Done()
	close(out)
	wg.Wait()
}

func handlerAddContact(conn *websocket.Conn, ctx context.Context, out chan<- any, cancel context.CancelFunc) {
	coll := db.Connect().Collection("user")

	for {
		var cont Contacts
		if err := conn.ReadJSON(&cont); err != nil {
			sendErr(ctx, out, "invalid request: "+err.Error())
			cancel()
			return
		}

		count, err := coll.CountDocuments(ctx, bson.M{
			"userId":           cont.UserID,
			"contact.friendId": cont.ContactID,
		})
		if err != nil {
			sendErr(ctx, out, err.Error())
			return
		}
		if count > 0 {
			sendErr(ctx, out, "contact already exists")
			continue
		}

		var user, contact Friends

		err = coll.FindOne(ctx, bson.M{
			"userId": cont.ContactID,
		}).Decode(&contact)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				sendErr(ctx, out, "contact does not exist")

			} else {
				sendErr(ctx, out, err.Error())
			}
			continue
		}
		err = coll.FindOne(ctx, bson.M{
			"userId": cont.UserID,
		}).Decode(&user)

		if err != nil {
			sendErr(ctx, out, err.Error())
			continue
		}

		session, err := db.Connect().Client().StartSession()
		if err != nil {
			sendErr(ctx, out, err.Error())
			return
		}

		_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
			if _, er := coll.UpdateOne(sc, bson.M{"userId": cont.UserID}, bson.M{
				"$push": bson.M{"contact": bson.M{"friendName": contact.Name, "friendId": cont.ContactID, "messages": bson.A{}}},
			}); er != nil {
				return nil, er
			}
			if _, er := coll.UpdateOne(sc, bson.M{"userId": cont.ContactID}, bson.M{
				"$push": bson.M{"contact": bson.M{"friendName": user.Name, "friendId": user.UserID, "messages": bson.A{}}},
			}); er != nil {
				return nil, er
			}

			return nil, nil
		})

		session.EndSession(ctx)

		if err != nil {
			sendErr(ctx, out, err.Error())
			continue
		}

		select {
		case out <- gin.H{
			"status": "success",
			"contact": gin.H{
				"id":   cont.ContactID,
				"name": contact.Name,
			},
		}:
		case <-ctx.Done():
			return
		}

	}
}

func sendErr(ctx context.Context, out chan<- any, msg string) {
	select {
	case out <- gin.H{utils.Err: msg}:
	case <-ctx.Done():
	}
}
