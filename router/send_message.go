package router

import (
	"fmt"
	"main/db"

	"main/upgrader"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Text     string `json:"text"`
	Date     string `json:"date"`
}

var update = make(chan Message)

func SendMessage(c *gin.Context) {
	u := upgrader.Upgrader()
	conn, err := u.Upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		conn.WriteJSON(gin.H{"err": err.Error()})
		fmt.Print(err.Error())
		return
	}

	go handlerSendMessage(conn, update)
	go handlerIncomingMessage(conn)
}

func handlerSendMessage(conn *websocket.Conn, update chan Message) {
	defer conn.Close()
	for {
		var messages Message
		if err := conn.ReadJSON(&messages); err != nil {
			conn.WriteJSON(gin.H{"err": err.Error()})
			break
		}
		filterSend := bson.M{
			"$and": bson.M{
				"userID":           messages.Sender,
				"contact.friendId": messages.Receiver,
			},
		}
		updateSend := bson.M{
			"contact.friendId": bson.M{
				"sender":   messages.Sender,
				"receiver": messages.Receiver,
				"dateTime": messages.Date,
				"text":     messages.Text,
			},
		}

		_, err := db.UpdateOne("user", filterSend, updateSend)
		if err != nil {
			conn.WriteJSON(gin.H{"err": err.Error()})
			break
		}

		filterRec := bson.M{
			"$and": bson.M{
				"userID":           messages.Receiver,
				"contact.friendId": messages.Sender,
			},
		}
		updateRec := bson.M{
			"contact.friendId": bson.M{
				"sender":   messages.Sender,
				"receiver": messages.Receiver,
				"dateTime": messages.Date,
				"text":     messages.Text,
			},
		}
		_, err = db.UpdateOne("user", filterRec, updateRec)
		if err != nil {
			conn.WriteJSON(gin.H{"err": err.Error()})
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

		update <- messages
	}
}
