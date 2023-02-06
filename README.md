# MetaTube SDK Go

[![Build Status](https://img.shields.io/github/actions/workflow/status/metatube-community/metatube-sdk-go/docker.yml?branch=main&style=flat-square&logo=github-actions)](https://github.com/metatube-community/metatube-sdk-go/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/metatube-community/metatube-sdk-go?style=flat-square)](https://github.com/metatube-community/metatube-sdk-go)
[![Require Go Version](https://img.shields.io/badge/go-%3E%3D1.20-30dff3?style=flat-square&logo=go)](https://github.com/metatube-community/metatube-sdk-go/blob/main/go.mod)
[![GitHub License](https://img.shields.io/github/license/metatube-community/metatube-sdk-go?color=A42E2B&logo=gnu&style=flat-square)](https://github.com/metatube-community/metatube-sdk-go/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/v/tag/metatube-community/metatube-sdk-go?color=%23ff8936&logo=fitbit&style=flat-square)](https://github.com/metatube-community/metatube-sdk-go/tags)

[//]: # ([![Supported Platforms]&#40;https://img.shields.io/badge/platform-Linux%20%7C%20FreeBSD%20%7C%20NetBSD%20%7C%20OpenBSD%20%7C%20Darwin%20%7C%20Windows-549688?style=flat-square&logo=launchpad&#41;]&#40;https://github.com/metatube-community/metatube-sdk-go&#41;)

Metadata Tube SDK in Golang.

## Contents

- [MetaTube SDK Go](#metatube-sdk-go)
	- [Contents](#contents)
    - [Features](#features)
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

## Features

- Supported platforms
  - Linux
  - Darwin
  - Windows
  - BSD(s)
- Supported Databases
  - [SQLite](https://gitlab.com/cznic/sqlite)
  - [PostgreSQL](https://github.com/jackc/pgx)
- Image processing
  - Auto cropping
  - Badge support
  - Face detection
  - Image hashing
- RESTful API
- 20+ providers
- Text translation

## Installation

To install this package, you first need [Go](https://golang.org/) installed (**version 1.20+ is required**), then you can use the below Go command to install SDK.

```sh
go get -u github.com/metatube-community/metatube-sdk-go
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

	"github.com/metatube-community/metatube-sdk-go/engine"
)

func main() {
	app := engine.Default()

	results, err := app.SearchMovieAll("<movie_id>", false)
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

You can find detailed examples in [examples folder](https://github.com/metatube-community/metatube-sdk-go/tree/main/_examples/) or specific implementations in [cmd folder](https://github.com/metatube-community/metatube-sdk-go/tree/main/cmd/).

### Initiate SDK engine manually

```go
package main

import (
	"log"
	"time"

	"github.com/metatube-community/metatube-sdk-go/database"
	"github.com/metatube-community/metatube-sdk-go/engine"
)

func main() {
	// Open database using in-memory SQLite.
	db, _ := database.Open(&database.Config{
		DSN:		  ":memory:",
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
	
	// Search actor from Xs/List with fallback enabled.
	app.SearchActor("<actor_name>", xslist.Name, true)
	
	// Search actor from all available providers with fallback enabled.
	app.SearchActorAll("<actor_name>", true)
	
	// Get actor metadata id from Xs/List with lazy enabled.
	app.GetActorInfoByProviderID(xslist.Name, "<id>", true)
	
	// Get actor metadata from given URL with lazy enabled.
	app.GetActorInfoByURL("https://<actor_page_url>", true)
}
```

### Search and get movie info

```go
func main() {
	app := engine.Default()
	
	// Search movie from JavBus with fallback enabled.
	app.SearchMovie("<movie_id>", javbus.Name, true)
	
	// Search movie from all available providers with fallback enabled.
	// Option fallback will search the database for movie info if the corresponding providers
	// fail to return valid metadata.
	app.SearchMovieAll("<movie_id>", true)
	
	// Get movie metadata id from ARZON with lazy enable.
	// With the lazy option set to true, it will first try to search the database and return
	// the info directly if it exists. If the lazy option is set to false, it will fetch info
	// from the given provider and update the database.
	app.GetMovieInfoByProviderID(arzon.Name, "<id>", true)
	
	// Get movie metadata from given URL with lazy enabled.
	app.GetMovieInfoByURL("https://<movie_page_url>", true)
}
```

### Get actor and movie images

```go
func main() {
	app := engine.Default()
	
	// Get actor primary image id from Xs/List.
	app.GetActorPrimaryImage(xslist.Name, "<id>")
	
	// Get movie primary image id from FANZA with aspect ratio and pos set to default.
	app.GetMoviePrimaryImage(fanza.Name, "<id>", -1, -1)
	
	// Get movie primary image id from FANZA with aspect ratio set to 7:10 and pos
	// set to the center.
	app.GetMoviePrimaryImage(fanza.Name, "<id>", 0.70, 0.5)
	
	// Get movie backdrop image id from SOD.
	app.GetMovieBackdropImage(sod.Name, "<id>")
}
```

### Text translate engine

```go
package main

import (
	"github.com/metatube-community/metatube-sdk-go/translate"
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

## Credits

| Library														                                           | Description																						                                                                    |
|-----------------------------------------------------------------|------------------------------------------------------------------------------------------------------|
| [gocolly/colly](https://github.com/gocolly/colly)			            | Elegant Scraper and Crawler Framework for Golang													                                        |
| [gin-gonic/gin](https://github.com/gin-gonic/gin)			            | Gin is a HTTP web framework written in Go															                                             |
| [gorm.io/gorm](https://gorm.io/)								                        | The fantastic ORM library for Golang																                                                 |
| [esimov/pigo](https://github.com/esimov/pigo)				               | Fast face detection, pupil/eyes localization and facial landmark points detection library in pure Go |
| [modernc.org/sqlite](https://gitlab.com/cznic/sqlite)		         | Package sqlite is a CGo-free port of SQLite/SQLite3												                                      |
| [corona10/goimagehash](https://github.com/corona10/goimagehash) | Go Perceptual image hashing package																                                                  |
| [antchfx/xpath](https://github.com/antchfx/xpath)			            | XPath package for Golang, supports HTML, XML, JSON document query									                           |

## License

[GNU GPLv3 License](https://github.com/metatube-community/metatube-sdk-go/blob/main/LICENSE)
