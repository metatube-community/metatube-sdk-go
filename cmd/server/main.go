package main

import (
	goflag "flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/peterbourgon/ff/v3"

	"github.com/metatube-community/metatube-sdk-go/database"
	"github.com/metatube-community/metatube-sdk-go/engine"
	V "github.com/metatube-community/metatube-sdk-go/internal/version"
	"github.com/metatube-community/metatube-sdk-go/route"
	"github.com/metatube-community/metatube-sdk-go/route/auth"
)

const defaultRequestTimeout = time.Minute

var (
	opts = new(options)
	flag = goflag.NewFlagSet("", goflag.ExitOnError)
)

type options struct {
	// main options
	bind  string
	port  string
	token string
	dsn   string

	// engine options
	requestTimeout time.Duration

	// database options
	dbMaxIdleConns int
	dbMaxOpenConns int
	dbAutoMigrate  bool
	dbPreparedStmt bool

	// version flag
	versionFlag bool
}

func init() {
	// gin initiate
	gin.DisableConsoleColor()

	// flag parsing
	flag.StringVar(&opts.bind, "bind", "", "Bind address of server")
	flag.StringVar(&opts.port, "port", "8080", "Port number of server")
	flag.StringVar(&opts.token, "token", "", "Token to access server")
	flag.StringVar(&opts.dsn, "dsn", "", "Database Service Name")
	flag.DurationVar(&opts.requestTimeout, "request-timeout", time.Minute, "Timeout per request")
	flag.IntVar(&opts.dbMaxIdleConns, "db-max-idle-conns", 0, "Database max idle connections")
	flag.IntVar(&opts.dbMaxOpenConns, "db-max-open-conns", 0, "Database max open connections")
	flag.BoolVar(&opts.dbAutoMigrate, "db-auto-migrate", false, "Database auto migration")
	flag.BoolVar(&opts.dbPreparedStmt, "db-prepared-stmt", false, "Database prepared statement")
	flag.BoolVar(&opts.versionFlag, "version", false, "Show version")
	ff.Parse(flag, os.Args[1:], ff.WithEnvVarNoPrefix())
}

func showVersionAndExit() {
	fmt.Println(V.VersionString())
	os.Exit(0)
}

func main() {
	if _, isSet := os.LookupEnv("VERSION"); opts.versionFlag &&
		!isSet /* NOTE: ignore this flag if ENV contains VERSION variable. */ {
		showVersionAndExit()
	}

	db, err := database.Open(&database.Config{
		DSN:                  opts.dsn,
		PreparedStmt:         opts.dbPreparedStmt,
		MaxIdleConns:         opts.dbMaxIdleConns,
		MaxOpenConns:         opts.dbMaxOpenConns,
		DisableAutomaticPing: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// always enable auto migrate for sqlite DB.
	if db.Config.Dialector.Name() == database.Sqlite {
		opts.dbAutoMigrate = true
	}

	// timeout must >= 1 second.
	if opts.requestTimeout < time.Second {
		opts.requestTimeout = defaultRequestTimeout
	}

	app := engine.New(db, opts.requestTimeout)
	if err = app.AutoMigrate(opts.dbAutoMigrate); err != nil {
		log.Fatal(err)
	}

	var token auth.Validator
	if opts.token != "" {
		token = auth.Token(opts.token)
	}

	var (
		addr   = net.JoinHostPort(opts.bind, opts.port)
		router = route.New(app, token)
	)
	if err = http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
