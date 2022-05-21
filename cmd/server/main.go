package main

import (
	goflag "flag"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/peterbourgon/ff/v3"

	"github.com/javtube/javtube-sdk-go/engine"
	"github.com/javtube/javtube-sdk-go/route"
	"github.com/javtube/javtube-sdk-go/route/validator"
)

var (
	opts = new(options)
	flag = goflag.NewFlagSet("", goflag.ExitOnError)
)

type options struct {
	bind        string
	port        string
	token       string
	dsn         string
	autoMigrate bool
}

func init() {
	// gin initiate
	gin.DisableConsoleColor()

	// flag parsing
	flag.StringVar(&opts.bind, "bind", "", "Bind address of server")
	flag.StringVar(&opts.port, "port", "8080", "Port number of server")
	flag.StringVar(&opts.token, "token", "", "Token to access server")
	flag.StringVar(&opts.dsn, "dsn", "", "Database Service Name")
	flag.BoolVar(&opts.autoMigrate, "auto-migrate", false, "Database auto migration")
	ff.Parse(flag, os.Args[1:], ff.WithEnvVarNoPrefix())
}

func main() {
	app, err := engine.New(&engine.Options{
		DSN:     opts.dsn,
		Timeout: 2 * time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err = app.AutoMigrate(opts.autoMigrate); err != nil {
		log.Fatal(err)
	}

	var token validator.Validator
	if opts.token != "" {
		token = validator.Token(opts.token)
	}

	var (
		addr   = net.JoinHostPort(opts.bind, opts.port)
		router = route.New(app, token)
	)
	if err = http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
