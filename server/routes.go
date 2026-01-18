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

	dagRouter := router.Group("/dag")
	{
		dagRouter.GET("/status", controller.GetDAGStatus)
		dagRouter.GET("/progress", controller.GetDAGProgress)
	}

	router.PUT("/:uuid/retry", controller.UpdateAndRetriggerJob)
	router.GET("/:uuid", controller.GetJobByUUID)
	router.GET("/", controller.GetJobs)
}
