package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/javtube/javtube-sdk-go/engine"
	"github.com/javtube/javtube-sdk-go/route"
)

func main() {
	gin.DisableConsoleColor()

	app, err := engine.New(&engine.Options{
		DSN:     "",
		Timeout: 30 * time.Second,
	})

	if err != nil {
		log.Fatal(err)
	}

	if err := app.AutoMigrate(); err != nil {
		log.Fatal(err)
	}

	router := route.New(app)

	if err := router.Run(); err != nil {
		log.Fatal(err)
	}
}
