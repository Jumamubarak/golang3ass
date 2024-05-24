package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"week4/internal/data"

	_ "github.com/lib/pq"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured by fetching enviroments. Err: %s", err)
	}

	var cfg config
	cfg.port, _ = strconv.Atoi(os.Getenv("PORT"))
	cfg.env = os.Getenv("ENV")
	cfg.db.dsn = os.Getenv("DB_DSN")
	cfg.db.maxOpenConns, _ = strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	cfg.db.maxIdleConns, _ = strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	cfg.db.maxIdleTime = os.Getenv("DB_MAX_IDLE_TIME")

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatalf("Connection failed. Error is: %s", err)
	}
	defer db.Close()

	logger.Printf("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	return db, nil
}
