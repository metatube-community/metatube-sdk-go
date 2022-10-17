# JavTube SDK Go

[![Build Status](https://img.shields.io/github/workflow/status/javtube/javtube-sdk-go/Publish%20Go%20Releases?style=flat-square&logo=github-actions)](https://github.com/javtube/javtube-sdk-go/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/javtube/javtube-sdk-go?style=flat-square)](https://github.com/javtube/javtube-sdk-go)
[![Require Go Version](https://img.shields.io/badge/go-%3E%3D1.19-30dff3?style=flat-square&logo=go)](https://github.com/javtube/javtube-sdk-go/blob/main/go.mod)
[![GitHub License](https://img.shields.io/github/license/javtube/javtube-sdk-go?color=A42E2B&logo=gnu&style=flat-square)](https://github.com/javtube/javtube-sdk-go/blob/main/LICENSE)
[![Supported Platforms](https://img.shields.io/badge/platform-Linux%20%7C%20FreeBSD%20%7C%20NetBSD%20%7C%20OpenBSD%20%7C%20Darwin%20%7C%20Windows-549688?style=flat-square&logo=launchpad)](https://github.com/javtube/javtube-sdk-go)
[![Tag](https://img.shields.io/github/v/tag/javtube/javtube-sdk-go?color=%23ff8936&logo=fitbit&style=flat-square)](https://github.com/javtube/javtube-sdk-go/tags)

Just Another Video Tube SDK in Golang.

## Contents

- [JavTube SDK Go](#javtube-sdk-go)
    - [Contents](#contents)
    - [Installation](#installation)
    - [Quickstart](#quickstart)
    - [API Examples](#api-examples)
        - [Initiate SDK engine manually](#initiate-sdk-engine-manually)
        - [Search and get actor info](#search-and-get-actor-info)
        - [Search and get movie info](#search-and-get-movie-info)
        - [Get actor and movie images](#get-actor-and-movie-images)
        - [Text translate engine](#text-translate-engine)
    - [License](#license)
    - [Credits](#credits)

## Installation

To install this package, you first need [Go](https://golang.org/) installed (**version 1.19+ is required**), then you can use the below Go command to install SDK.

```sh
go get -u github.com/javtube/javtube-sdk-go
```

## Quickstart

```sh
# assume the following codes in example.go file
$ cat example.go
```

```go
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
```

```sh
# run example.go and see output on console
$ go run example.go
```

## API Examples

You can find detailed examples in [examples folder](https://github.com/javtube/javtube-sdk-go/tree/main/_examples/) or specific implementations in [cmd folder](https://github.com/javtube/javtube-sdk-go/tree/main/cmd/).

### Initiate SDK engine manually

```go
package main

import (
	"log"
	"time"

	"github.com/javtube/javtube-sdk-go/database"
	"github.com/javtube/javtube-sdk-go/engine"
)

func main() {
	// Open database using in-memory SQLite.
	db, _ := database.Open(&database.Config{
		DSN:          ":memory:",
		PreparedStmt: false,
	})

	// Allocate app engine with request timeout set to one minute.
	app := engine.New(db, time.Minute)

	// Initiate DB tables, only required at the first time.
	if err := app.AutoMigrate(true); err != nil {
		log.Fatal(err)
	}
}
```

### Search and get actor info

```go
func main() {
    app := engine.Default()
    
    // Search actor named `ひなたまりん` from Xs/List with fallback enabled.
    app.SearchActor("ひなたまりん", xslist.Name, true)
    
    // Search actor named `一ノ瀬もも` from all available providers with fallback enabled.
    app.SearchActorAll("一ノ瀬もも", true)
    
    // Get actor metadata id `5085` from Xs/List with lazy enabled.
    app.GetActorInfoByProviderID(xslist.Name, "5085", true)
    
    // Get actor metadata from given URL with lazy enabled.
    app.GetActorInfoByURL("https://xslist.org/zh/model/15659.html", true)
}
```

### Search and get movie info

```go
func main() {
    app := engine.Default()
    
    // Search movie named `ABP-330` from JavBus with fallback enabled.
    app.SearchMovie("ABP-330", javbus.Name, true)
    
    // Search movie named `SSIS-110` from all available providers with fallback enabled.
    // Option fallback will search the database for movie info if the corresponding providers
    // fail to return valid metadata.
    app.SearchMovieAll("SSIS-110", true)
    
    // Get movie metadata id `1252925` from ARZON with lazy enable.
    // With the lazy option set to true, it will first try to search the database and return
    // the info directly if it exists. If the lazy option is set to false, it will fetch info
    // from the given provider and update the database.
    app.GetMovieInfoByProviderID(arzon.Name, "1252925", true)
    
    // Get movie metadata from given URL with lazy enabled.
    app.GetMovieInfoByURL("https://www.heyzo.com/moviepages/2189/index.html", true)
}
```

### Get actor and movie images

```go
func main() {
    app := engine.Default()
    
    // Get actor primary image id `24490` from Xs/List.
    app.GetActorPrimaryImage(xslist.Name, "24490")
    
    // Get movie primary image id `hmn00268` from FANZA with aspect ratio and pos set to default.
    app.GetMoviePrimaryImage(fanza.Name, "hmn00268", -1, -1)
    
    // Get movie primary image id `hmn00268` from FANZA with aspect ratio set to 7:10 and pos set to center.
    app.GetMoviePrimaryImage(fanza.Name, "hmn00268", 0.70, 0.5)
    
    // Get movie backdrop image id `DLDSS-077` from SOD.
    app.GetMovieBackdropImage(sod.Name, "DLDSS-077")
}
```

### Text translate engine

```go
package main

import (
	"github.com/javtube/javtube-sdk-go/translate"
)

func main() {
	var (
		appId  = "XXX"
		appKey = "XXX"
	)
	// Translate `Hello` from auto to Japanese by Baidu.
	translate.BaiduTranslate("Hello", "auto", "ja", appId, appKey)

	var apiKey = "XXX"
	// Translate `Hello` from auto to simplified Chinese by Google.
	translate.GoogleTranslate("Hello", "auto", "zh-cn", apiKey)
}
```

## License

This project is opened under the [GNU GPLv3](https://github.com/javtube/javtube-sdk-go/blob/main/LICENSE) license.

## Credits

| Library                                                         | Description                                                                                          |
|-----------------------------------------------------------------|------------------------------------------------------------------------------------------------------|
| [gocolly/colly](https://github.com/gocolly/colly)               | Elegant Scraper and Crawler Framework for Golang                                                     |
| [gin-gonic/gin](https://github.com/gin-gonic/gin)               | Gin is a HTTP web framework written in Go                                                            |
| [gorm.io/gorm](https://gorm.io/)                                | The fantastic ORM library for Golang                                                                 |
| [esimov/pigo](https://github.com/esimov/pigo)                   | Fast face detection, pupil/eyes localization and facial landmark points detection library in pure Go |
| [modernc.org/sqlite](https://gitlab.com/cznic/sqlite)           | Package sqlite is a CGo-free port of SQLite/SQLite3                                                  |
| [corona10/goimagehash](https://github.com/corona10/goimagehash) | Go Perceptual image hashing package                                                                  |
| [antchfx/xpath](https://github.com/antchfx/xpath)               | XPath package for Golang, supports HTML, XML, JSON document query                                    |
