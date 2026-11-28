# MetaTube SDK Go

[![Build Status](https://img.shields.io/github/actions/workflow/status/metatube-community/metatube-sdk-go/docker.yml?branch=main&style=flat-square&logo=githubactions)](https://github.com/metatube-community/metatube-sdk-go/actions/workflows/docker.yml)
[![Go Linter](https://img.shields.io/github/actions/workflow/status/metatube-community/metatube-sdk-go/linter.yml?style=flat-square&branch=main&logo=githubactions&label=golangci-lint)](https://github.com/metatube-community/metatube-sdk-go/actions/workflows/linter.yml)
[![Require Go Version](https://img.shields.io/badge/go-%3E%3D1.27-00ADD8?style=flat-square&logo=go&logoColor=white)](https://github.com/metatube-community/metatube-sdk-go/blob/main/go.mod)
[![GitHub License](https://img.shields.io/github/license/metatube-community/metatube-sdk-go?color=e4682a&logo=apache&style=flat-square)](https://github.com/metatube-community/metatube-sdk-go/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/v/tag/metatube-community/metatube-sdk-go?color=%23ff8936&logo=git&logoColor=white&style=flat-square)](https://github.com/metatube-community/metatube-sdk-go/tags)

MetaTube SDK and API server in Golang. See the [documentation](https://metatube-community.github.io) for setup.

## Features

- Supported platforms
    - Linux
    - Darwin
    - Windows
    - BSD(s)
- Supported databases
    - [SQLite](https://gitlab.com/cznic/sqlite)
    - [PostgreSQL](https://github.com/jackc/pgx)
- Image processing
    - Auto cropping
    - Badge support
    - Face detection
    - Image hashing
- RESTful API
- 30+ providers
- Text translation

`JavDB` provides basic FC2 metadata from public search results as a lower-priority
alternative to `FC2`, `fc2hub`, and `FC2PPVDB`. It matches the exact FC2 number and
provides the title, cover, release date, and rating without requiring a login.
Actor names and other detail-page fields are not available through this provider.
Its IDs are numeric FC2 IDs; supported URLs are public search URLs such as
`https://javdb.com/search?q=FC2-3119569&f=all`.

## Installation

To install the SDK, you need [Go](https://go.dev/) installed (**go1.27+ is required**), then run:

```sh
go get -u github.com/metatube-community/metatube-sdk-go
```

## Credits

| Library                                                         | Description                                                                                          |
|-----------------------------------------------------------------|------------------------------------------------------------------------------------------------------|
| [gocolly/colly](https://github.com/gocolly/colly)               | Elegant Scraper and Crawler Framework for Golang                                                     |
| [gin-gonic/gin](https://github.com/gin-gonic/gin)               | Gin is a HTTP web framework written in Go                                                            |
| [gorm.io/gorm](https://gorm.io/)                                | The fantastic ORM library for Golang                                                                 |
| [esimov/pigo](https://github.com/esimov/pigo)                   | Fast face detection, pupil/eyes localization and facial landmark points detection library in pure Go |
| [robertkrimen/otto](https://github.com/robertkrimen/otto)       | A JavaScript interpreter in Go (golang)                                                              |
| [modernc.org/sqlite](https://gitlab.com/cznic/sqlite)           | Package sqlite is a CGo-free port of SQLite/SQLite3                                                  |
| [corona10/goimagehash](https://github.com/corona10/goimagehash) | Go Perceptual image hashing package                                                                  |
| [antchfx/htmlquery](https://github.com/antchfx/htmlquery)       | htmlquery is golang XPath package for HTML query                                                     |
| [gen2brain/jpegli](https://github.com/gen2brain/jpegli)         | Go encoder/decoder for JPEG based on jpegli                                                          |

## License

[Apache-2.0 License](https://github.com/metatube-community/metatube-sdk-go/blob/main/LICENSE)
