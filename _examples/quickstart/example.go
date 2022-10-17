package main

import (
	"fmt"
	"log"

	"github.com/javtube/javtube-sdk-go/engine"
)

func main() {
	app := engine.Default()

	results, err := app.SearchMovieAll("GVH-466", false)
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range results {
		fmt.Println(result.Provider, result.ID, result.Number, result.Title)
	}
}
