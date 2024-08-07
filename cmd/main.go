package cmd

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

var config = &struct {
	// main config
	bind  string
	port  string
	token string
	dsn   string

	// engine config
	requestTimeout time.Duration

	// database config
	dbMaxIdleConns int
	dbMaxOpenConns int
	dbAutoMigrate  bool
	dbPreparedStmt bool

	// version flag
	versionFlag bool
}{}

func init() {
	// gin init
	gin.DisableConsoleColor()

	// flag init
	flag := goflag.NewFlagSet("", goflag.ExitOnError)

	// flag parsing
	flag.StringVar(&config.bind, "bind", "", "Bind address of server")
	flag.StringVar(&config.port, "port", "8080", "Port number of server")
	flag.StringVar(&config.token, "token", "", "Token to access server")
	flag.StringVar(&config.dsn, "dsn", "", "Database Service Name")
	flag.DurationVar(&config.requestTimeout, "request-timeout", engine.DefaultRequestTimeout, "Timeout per request")
	flag.IntVar(&config.dbMaxIdleConns, "db-max-idle-conns", 0, "Database max idle connections")
	flag.IntVar(&config.dbMaxOpenConns, "db-max-open-conns", 0, "Database max open connections")
	flag.BoolVar(&config.dbAutoMigrate, "db-auto-migrate", false, "Database auto migration")
	flag.BoolVar(&config.dbPreparedStmt, "db-prepared-stmt", false, "Database prepared statement")
	flag.BoolVar(&config.versionFlag, "version", false, "Show version")
	ff.Parse(flag, os.Args[1:], ff.WithEnvVars())
}

func showVersionAndExit() {
	fmt.Println(V.BuildString())
	os.Exit(0)
}

func Router(names ...string) *gin.Engine {
	db, err := database.Open(&database.Config{
		DSN:                  config.dsn,
		PreparedStmt:         config.dbPreparedStmt,
		MaxIdleConns:         config.dbMaxIdleConns,
		MaxOpenConns:         config.dbMaxOpenConns,
		DisableAutomaticPing: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// engine options
	var opts []engine.Option

	// timeout must >= 1 second
	if config.requestTimeout >= time.Second {
		opts = append(opts, engine.WithRequestTimeout(config.requestTimeout))
	}

	// specify engine name
	for _, name := range names {
		opts = append(opts, engine.WithEngineName(name))
	}

	app := engine.New(db, opts...)

	// always enable auto migrate for sqlite DB
	if app.DBType() == database.Sqlite {
		config.dbAutoMigrate = true
	}
	if err = app.DBAutoMigrate(config.dbAutoMigrate); err != nil {
		log.Fatal(err)
	}

	var token auth.Validator
	if config.token != "" {
		token = auth.Token(config.token)
	}

	return route.New(app, token)
}

func Main() {
	if _, isSet := os.LookupEnv("VERSION"); config.versionFlag &&
		!isSet /* NOTE: ignore this flag if ENV contains VERSION variable. */ {
		showVersionAndExit()
	}

	var (
		addr   = net.JoinHostPort(config.bind, config.port)
		router = Router(engine.DefaultEngineName)
	)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
