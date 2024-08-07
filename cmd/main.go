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

var (
	cfg  = new(config)
	flag = goflag.NewFlagSet("", goflag.ExitOnError)
)

type config struct {
	// main config
	bind  string
	port  string
	name  string
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
}

func init() {
	// gin init
	gin.DisableConsoleColor()

	// flag parsing
	flag.StringVar(&cfg.bind, "bind", "", "Bind address of server")
	flag.StringVar(&cfg.port, "port", "8080", "Port number of server")
	flag.StringVar(&cfg.token, "token", "", "Token to access server")
	flag.StringVar(&cfg.dsn, "dsn", "", "Database Service Name")
	flag.StringVar(&cfg.name, "name", engine.DefaultEngineName, "Application name of server")
	flag.DurationVar(&cfg.requestTimeout, "request-timeout", engine.DefaultRequestTimeout, "Timeout per request")
	flag.IntVar(&cfg.dbMaxIdleConns, "db-max-idle-conns", 0, "Database max idle connections")
	flag.IntVar(&cfg.dbMaxOpenConns, "db-max-open-conns", 0, "Database max open connections")
	flag.BoolVar(&cfg.dbAutoMigrate, "db-auto-migrate", false, "Database auto migration")
	flag.BoolVar(&cfg.dbPreparedStmt, "db-prepared-stmt", false, "Database prepared statement")
	flag.BoolVar(&cfg.versionFlag, "version", false, "Show version")
	ff.Parse(flag, os.Args[1:], ff.WithEnvVars())
}

func showVersionAndExit() {
	fmt.Println(V.BuildString())
	os.Exit(0)
}

func Router() *gin.Engine {
	db, err := database.Open(&database.Config{
		DSN:                  cfg.dsn,
		PreparedStmt:         cfg.dbPreparedStmt,
		MaxIdleConns:         cfg.dbMaxIdleConns,
		MaxOpenConns:         cfg.dbMaxOpenConns,
		DisableAutomaticPing: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// engine options
	var opts []engine.Option

	// timeout must >= 1 second
	if cfg.requestTimeout >= time.Second {
		opts = append(opts, engine.WithRequestTimeout(cfg.requestTimeout))
	}

	// specify engine name
	if cfg.name != "" {
		opts = append(opts, engine.WithEngineName(cfg.name))
	}

	app := engine.New(db, opts...)

	// always enable auto migrate for sqlite DB
	if app.DBType() == database.Sqlite {
		cfg.dbAutoMigrate = true
	}
	if err = app.DBAutoMigrate(cfg.dbAutoMigrate); err != nil {
		log.Fatal(err)
	}

	var token auth.Validator
	if cfg.token != "" {
		token = auth.Token(cfg.token)
	}

	return route.New(app, token)
}

func Main() {
	if _, isSet := os.LookupEnv("VERSION"); cfg.versionFlag &&
		!isSet /* NOTE: ignore this flag if ENV contains VERSION variable. */ {
		showVersionAndExit()
	}

	var (
		addr   = net.JoinHostPort(cfg.bind, cfg.port)
		router = Router()
	)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
