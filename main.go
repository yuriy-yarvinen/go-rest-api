package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()
	server.GET("/ping", func(ginCtx *gin.Context) {
		ginCtx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	server.GET("/events", getEvents) // Register the /events route with the getEvents handler
	server.Run(":8081")              // listen and serve on 0.0.0.0:8081
}

func getEvents(context *gin.Context) {
	// Implement your logic to retrieve events here
	events := []string{"Event 1", "Event 2", "Event 3"} // Example events

	context.JSON(http.StatusOK, gin.H{
		"events": events,
	})
}
