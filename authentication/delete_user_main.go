package authentication

import (
	"main/db"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func DeleteUserMain(c *gin.Context) {
	var req Req

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	filter := bson.M{
		req.Name: bson.M{
			"$exists": true,
		},
	}
	var userDoc map[string]any

	if err := db.FindONe("main", filter, &userDoc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	if _, err := db.DeleteOne("main", bson.M{"_id": userDoc["_id"]}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success delete": userDoc["_id"]})
}
