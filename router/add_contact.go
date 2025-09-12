// Package router
package router

import (
	"fmt"
	"log"
	"main/db"
	"main/upgrader"

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
	UserID string `bson:"userID"`
}

func AddCountact(c *gin.Context) {
	u := upgrader.Upgrader()
	conn, err := u.Upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		conn.WriteJSON(gin.H{"err": err.Error()})
		fmt.Print(err.Error())
		return
	}

	go handlerAddContact(conn)

}

func handlerAddContact(conn *websocket.Conn) {
	defer conn.Close()

	for {
		var cont Contacts
		if errR := conn.ReadJSON(&cont); errR != nil {
			conn.WriteJSON(gin.H{"error when reading": errR.Error()})
			fmt.Println("error when reading :", errR.Error())
			break
		}

		friend, err := db.CheckDoc("user", bson.M{
			"userID":           cont.UserID,
			"contact.friendId": cont.ContactID,
		})
		if err != nil {
			log.Println(err.Error())
			conn.WriteJSON(gin.H{"error when chek contact": err.Error()})
			break
		}
		if friend > 0 {
			conn.WriteJSON(bson.M{"err": "has exist contact"})
			break
		}
		user := Friends{}
		contact := Friends{}
		err = db.FindONe("user", bson.M{
			"userID": cont.ContactID,
		}, &contact)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				conn.WriteJSON(gin.H{"err": "contact doesn't exists "})
				break
			} else {
				conn.WriteJSON(gin.H{"err": err.Error()})
				break
			}
		}
		err = db.FindONe("user", bson.M{
			"userID": cont.UserID,
		}, &user)
		if err != nil {
			conn.WriteJSON(gin.H{"err": err.Error()})
			break
		}

		updateUser := bson.M{
			"$push": bson.M{
				"contact": bson.M{
					"name":        contact.Name,
					"participant": contact.UserID,
					"friendId":    cont.ContactID,
					"messages":    bson.A{},
				},
			},
		}
		updateCont := bson.M{
			"$push": bson.M{
				"contact": bson.M{
					"name":        contact.Name,
					"participant": contact.UserID,
					"friendId":    cont.ContactID,
					"messages":    bson.A{},
				},
			},
		}

		_, errs := db.UpdateOne("user", bson.M{"userID": cont.UserID}, updateUser)
		_, erru := db.UpdateOne("user", bson.M{"userID": cont.ContactID}, updateCont)
		if errs != nil {
			conn.WriteJSON(gin.H{"err": errs.Error()})
			break
		}
		if erru != nil {
			conn.WriteJSON(gin.H{"err": erru.Error()})
			return
		}
		if err = conn.WriteJSON(bson.M{"success": cont.ContactID}); err != nil {
			log.Println(err.Error())
			break
		}
	}
}
