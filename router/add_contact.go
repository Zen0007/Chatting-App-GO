// Package router
package router

import (
	"context"
	"fmt"
	"log"
	"main/db"
	"main/upgrader"
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

func AddCountact(c *gin.Context) {
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
		defer wg.Done()
		defer conn.Close()
		for msg := range out {
			if err := conn.WriteJSON(msg); err != nil {
				log.Println("error when write", err.Error())
				cancel()
				break
			}
		}
	}()

	wg.Add(1)
	go func() {
		wg.Done()
		handlerAddContact(conn, ctx, out)
	}()
	// Wait until connection is closed
	<-ctx.Done()
	wg.Wait()
	close(out)

}

func handlerAddContact(conn *websocket.Conn, ctx context.Context, out chan<- any) {
	for {
		var cont Contacts
		if errR := conn.ReadJSON(&cont); errR != nil {
			select {
			case out <- gin.H{"error when reading": errR.Error()}:
			case <-ctx.Done():
			}
			fmt.Println("error when reading :", errR.Error())
			break
		}

		friend, err := db.CheckDoc("user", bson.M{
			"userId":           cont.UserID,
			"contact.friendId": cont.ContactID,
		})
		if err != nil {
			log.Println(err.Error())
			select {
			case out <- gin.H{"error when chek contact": err.Error()}:
			case <-ctx.Done():
			}
			break
		}
		if friend > 0 {
			select {
			case out <- bson.M{"err": "has exist contact"}:
			case <-ctx.Done():
			}
			break
		}
		user := Friends{}
		contact := Friends{}
		err = db.FindONe("user", bson.M{
			"userId": cont.ContactID,
		}, &contact)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				select {
				case out <- gin.H{"err": "contact doesn't exists "}:
				case <-ctx.Done():
				}
				break
			} else {
				select {
				case out <- gin.H{"err": err.Error()}:
				case <-ctx.Done():
				}
				break
			}
		}
		err = db.FindONe("user", bson.M{
			"userId": cont.UserID,
		}, &user)
		if err != nil {
			select {
			case out <- gin.H{"err": err.Error()}:
			case <-ctx.Done():
			}
			break
		}

		updateUser := bson.M{
			"$push": bson.M{
				"contact": bson.M{
					"frienName": contact.Name,
					"friendId":  cont.ContactID,
					"messages":  bson.A{},
				},
			},
		}
		updateCont := bson.M{
			"$push": bson.M{
				"contact": bson.M{
					"frienName": user.Name,
					"friendId":  user.UserID,
					"messages":  bson.A{},
				},
			},
		}

		_, errs := db.UpdateOne("user", bson.M{"userId": cont.UserID}, updateUser)
		_, erru := db.UpdateOne("user", bson.M{"userId": cont.ContactID}, updateCont)
		if errs != nil {
			select {
			case out <- gin.H{"err": errs.Error()}:
			case <-ctx.Done():
			}
			break
		}
		if erru != nil {
			select {
			case out <- gin.H{"err": erru.Error()}:
			case <-ctx.Done():
			}
			break
		}

		select {
		case out <- gin.H{"success": cont.ContactID}:
		case <-ctx.Done():
		}
	}
}
