package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/namsral/flag"
	"github.com/twitchscience/gologging/gologging"
)

// Holds application-level config
type AppConfig struct {
	debug       bool
	host        string
	port        int
	num_workers int
	logging_dir string
	aws_region  string
	bucket      string
	key_prefix  string
}

type AppContext struct {
	config *AppConfig

	// global application logger
	logger *log.Logger

	// S3 event logger
	s3log *gologging.UploadLogger
}

func main() {
	config := AppConfig{}
	app := AppContext{config: &config}

	flag.BoolVar(&config.debug, "debug", false, "Debug mode: log to stderr instead of S3")
	flag.StringVar(&config.host, "host", "0.0.0.0", "Host to bind HTTP server on")
	flag.IntVar(&config.port, "port", 3000, "Port to bind HTTP server on")
	flag.IntVar(&config.num_workers, "num-workers", 1, "Number of workers uploading logs")
	flag.StringVar(&config.logging_dir, "logging-dir", ".", "Directory to store temp log files")
	flag.StringVar(&config.aws_region, "aws-region", "us-west-1", "AWS region")
	flag.StringVar(&config.bucket, "bucket", "logs", "S3 bucket for storing logs")
	flag.StringVar(&config.key_prefix, "key-prefix", "", "Prefix for S3 keys")
	flag.Parse()

	app.logger = log.New(os.Stderr, "[metrics] ", log.LstdFlags|log.Lshortfile)

	s3log, err := app.setupS3Logger()
	if err != nil {
		app.logger.Fatal("failed starting S3 logger", err)
	}
	app.s3log = s3log

	router := app.setupRoutes()

	// Make sure logger is flushed when shutdown signal is received
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		log.Println("interrupted, closing logger...")
		s3log.Close()
		os.Exit(0)
	}()

	// Start the server!
	hostname := fmt.Sprintf("%s:%d", config.host, config.port)
	app.logger.Println("http server listening on", hostname)
	if err := http.ListenAndServe(hostname, router); err != nil {
		app.logger.Fatal("error starting HTTP server:", err)
	}
}
