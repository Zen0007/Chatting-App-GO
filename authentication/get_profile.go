package authentication

import (
	"main/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetProfile(c *gin.Context) {
	var req Req
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	filter := bson.M{
		"email":    req.Email,
		"username": req.Name,
	}
	var userDoc map[string]any

	if err := db.FindONe("account", filter, &userDoc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": userDoc})
}
