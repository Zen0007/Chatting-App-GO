package authentication

import (
	"main/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Contact struct {
	ID      primitive.ObjectID `bson:"_id"`
	Contact []string           `bson:"contact"`
}

func ContactFriends(c *gin.Context) {
	var contact Contact
	filter := bson.M{
		"contacts": bson.M{
			"$exists": true,
		},
	}
	if err := db.FindONe("contact_friends", filter, &contact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": contact})
}
