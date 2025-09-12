// Package authentication
package authentication

import (
	"fmt"
	"log"
	"main/db"
	"main/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type Auth struct {
	Name     string `bson:"name,omitempty" json:"name,omitempty"`
	Email    string `bson:"email" json:"email"`
	Password string `bson:"password" json:"password"`
}

func Register(c *gin.Context) {
	var user Auth
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		log.Fatal(err.Error())
		return
	}

	if len(user.Name) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"err": "name too short"})
		return
	}

	if len(user.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"err": "password too short"})
		return
	}
	count, err := db.CheckDoc("account", bson.M{"email": user.Email})
	if err != nil {
		log.Fatal(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"err": "email has register please login "})
		return
	}

	user.Password, err = utils.GenerateHashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": "could not generate password has"})
		return
	}

	userID := utils.GenerateIDUser(10)
	ms := bson.M{
		"userID":   userID,
		"userName": user.Name,
		"email":    user.Email,
		"password": user.Password,
		"contact":  bson.A{},
	}
	_, err = db.InsertOne("user", ms)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("database crash: %s", err.Error())})
		return
	}

	keys := bson.M{
		"userID":   userID,
		"username": user.Name,
		"email":    user.Email,
		"password": user.Password,
	}

	_, errd := db.InsertOne("account", keys)
	if errd != nil {
		fmt.Println(errd.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"err": fmt.Sprintf("database crash: %s", errd.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": gin.H{
			"userId":   userID,
			"userName": user.Name,
			"email":    user.Email,
		},
	})

}
