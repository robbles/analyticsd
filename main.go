package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/namsral/flag"

	_ "github.com/go-sql-driver/mysql"
)

// Holds application-level config
type AppConfig struct {
	debug        bool
	database_url string
	host         string
	port         int
}

type AppContext struct {
	config *AppConfig

	// global application logger
	logger *log.Logger

	// This is a concurrency-safe database connection pool
	db *sqlx.DB
}

func main() {
	config := AppConfig{}
	app := AppContext{config: &config}

	flag.BoolVar(&config.debug, "debug", false, "Debug mode")
	flag.StringVar(&config.database_url, "database-url", "/",
		"Database connection string: [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]")
	flag.StringVar(&config.host, "host", "0.0.0.0", "Host to bind HTTP server on")
	flag.IntVar(&config.port, "port", 3000, "Port to bind HTTP server on")
	flag.Parse()

	app.logger = log.New(os.Stderr, "[ad_server] ", log.LstdFlags|log.Lshortfile)

	app.db = sqlx.MustConnect("mysql", config.database_url)

	router := setupRoutes(&app)

	hostname := fmt.Sprintf("%s:%d", config.host, config.port)

	// Start the server!
	app.logger.Println("http server listening on", hostname)
	if err := http.ListenAndServe(hostname, router); err != nil {
		app.logger.Fatal("error starting HTTP server:", err)
	}
}
