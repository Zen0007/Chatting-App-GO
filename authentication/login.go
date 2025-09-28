package authentication

import (
	"main/db"
	"main/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginUser(c *gin.Context) {
	var login Login

	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(400, gin.H{utils.Err: err.Error()})
		return
	}
	var user Login
	err := db.FindOne("user", bson.M{"email": login.Email}, &user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(400, gin.H{utils.Err: "user not exist"})
			return
		} else {
			c.JSON(500, gin.H{utils.Err: "internal server error"})
			return
		}

	}
	errHash := utils.CompareHashPassword(login.Password, user.Password)
	if errHash != nil {
		c.JSON(400, gin.H{utils.Err: "invalid password", "e": errHash.Error()})
		return
	}

	tokenString, err := utils.ParseToken(user.Email)
	if err != nil {
		c.JSON(400, gin.H{utils.Err: err.Error()})
		return
	}

	c.SetCookie("token", tokenString, int(time.Now().Add(time.Hour*32).Unix()), "/", "localhost", false, true)
	c.JSON(200, gin.H{utils.Success: "user logged in"})
}
