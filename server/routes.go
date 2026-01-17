package server

import (
	"vivek-ray/server/controller"

	"github.com/gin-gonic/gin"
)

func Routes(router *gin.RouterGroup) {
	bulkInsertRouter := router.Group("/bulk-insert")
	{
		bulkInsertRouter.POST("/complete-graph", controller.BulkInsertCompleteGraph)
	}
	router.PUT("/:uuid/retry", controller.UpdateAndRetriggerJob)
	router.GET("/", controller.GetJobs)
}
