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
		c.JSON(http.StatusBadRequest, gin.H{utils.Err: err.Error()})
		log.Fatal(err.Error())
		return
	}

	if len(user.Name) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{utils.Err: "name too short"})
		return
	}

	if len(user.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{utils.Err: "password too short"})
		return
	}
	count, err := db.CheckDoc("user", bson.M{"email": user.Email})
	if err != nil {
		log.Fatal(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{utils.Err: err.Error()})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{utils.Err: "email has register please login "})
		return
	}

	user.Password, err = utils.GenerateHashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{utils.Err: "could not generate password has"})
		return
	}

	userID := fmt.Sprint("0007", utils.GenerateIDUser(5))
	ms := bson.M{
		"userId":   userID,
		"userName": user.Name,
		"email":    user.Email,
		"password": user.Password,
		"contact":  bson.A{},
	}
	_, err = db.InsertOne("user", ms)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{utils.Err: fmt.Sprintf("database crash: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		utils.Success: gin.H{
			"userId":   userID,
			"userName": user.Name,
			"email":    user.Email,
		},
	})

}
