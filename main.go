package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pcittadini/gotick/handlers/taskHandlers"
)

func main() {

	// gotick consumers
	//go clients.TestConsumer(1)
	//go clients.TestConsumer(2)

	// gotick scheduler
	//go tasks.Scheduler()

	router := gin.Default()

	// Simple group: v1
	taskAPI := router.Group("/tasks")
	{
		taskAPI.POST("/new", taskHandlers.New)
	}

	router.Run(":8080")

}

