package controller

import (
	"net/http"
	"vivek-ray/connections"
	"vivek-ray/server/services"

	"github.com/gin-gonic/gin"
)

func GetDAGStatus(c *gin.Context) {
	uuids := c.QueryArray("uuids")
	if len(uuids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one uuid is required"})
		return
	}

	dagStatus, err := services.NewDAGService(connections.PgDBConnection.Client).GetDAGStatus(uuids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dagStatus)
}

func GetDAGProgress(c *gin.Context) {
	uuids := c.QueryArray("uuids")
	if len(uuids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one uuid is required"})
		return
	}

	progress, err := services.NewDAGService(connections.PgDBConnection.Client).GetDAGProgress(uuids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}
