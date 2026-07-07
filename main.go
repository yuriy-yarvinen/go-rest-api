package main

import (
	"go-rest-api/database"
	"go-rest-api/events"
	"go-rest-api/events/postgres"
	"go-rest-api/events/rest"
	"go-rest-api/events/sqlite"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()
	database.CreateTables()

	// Wire the layers: infrastructure -> application -> transport.
	// The concrete repository is picked based on the driver InitDB connected to.
	var eventRepo events.EventRepository
	switch database.Driver {
	case "postgres":
		eventRepo = postgres.NewRepository(database.DB)
	default:
		eventRepo = sqlite.NewRepository(database.DB)
	}
	eventService := events.NewService(eventRepo)
	eventHandler := rest.NewHandler(eventService)

	server := gin.Default()
	rest.RegisterRoutes(server.Group("/api/v1/"), eventHandler)
	server.Run(":8082") // listen and serve on 0.0.0.0:8082
}
