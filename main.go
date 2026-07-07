package main

import (
	"go-rest-api/database"
	"go-rest-api/events"
	eventspostgres "go-rest-api/events/postgres"
	eventsrest "go-rest-api/events/rest"
	eventssqlite "go-rest-api/events/sqlite"
	"go-rest-api/users"
	userspostgres "go-rest-api/users/postgres"
	usersrest "go-rest-api/users/rest"
	userssqlite "go-rest-api/users/sqlite"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()
	database.CreateTables()

	// Wire the layers: infrastructure -> application -> transport.
	// The concrete repositories are picked based on the driver InitDB connected to.
	var eventRepo events.EventRepository
	var userRepo users.UserRepository
	switch database.Driver {
	case "postgres":
		eventRepo = eventspostgres.NewRepository(database.DB)
		userRepo = userspostgres.NewRepository(database.DB)
	default:
		eventRepo = eventssqlite.NewRepository(database.DB)
		userRepo = userssqlite.NewRepository(database.DB)
	}

	eventService := events.NewService(eventRepo)
	eventHandler := eventsrest.NewHandler(eventService)

	userService := users.NewService(userRepo)
	userHandler := usersrest.NewHandler(userService)

	server := gin.Default()
	api := server.Group("/api/v1/")
	eventsrest.RegisterRoutes(api, eventHandler)
	usersrest.RegisterRoutes(api, userHandler)
	server.Run(":8082") // listen and serve on 0.0.0.0:8082
}
