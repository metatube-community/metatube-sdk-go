package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/javtube/javtube-sdk-go/engine"
	"github.com/javtube/javtube-sdk-go/route"
	"github.com/javtube/javtube-sdk-go/route/validator"
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

	token := validator.Token("token")

	router := route.New(app, token)

	if err := router.Run(); err != nil {
		log.Fatal(err)
	}
}
