package authentication

import (
	"main/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Req struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

type Profile struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name,omitempty" json:"name,omitempty"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"password"`
}

func DeleteUserProfile(c *gin.Context) {
	var user Req
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	var data Profile
	filter := bson.M{
		"email": user.Email,
	}
	_, err := db.Find("user_profile", filter, &data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	filter = bson.M{
		"email": data.Email,
	}
	if _, err := db.DeleteOne("user_profile", filter); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": data.ID})
}
