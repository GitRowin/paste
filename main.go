package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

//go:embed assets/*
var assets embed.FS

type App struct {
	db *sql.DB
}

func main() {
	db, err := sql.Open("sqlite3", "paste.db")

	if err != nil {
		panic(err)
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL")

	if err != nil {
		panic(err)
	}

	query := `CREATE TABLE IF NOT EXISTS pastes(
	id TEXT PRIMARY KEY,
	created_at INTEGER NOT NULL,
	country_code TEXT NOT NULL,
	views INTEGER NOT NULL DEFAULT 0,
	last_view INTEGER DEFAULT NULL,
	content BLOB NOT NULL
) STRICT`

	_, err = db.Exec(query)

	if err != nil {
		panic(err)
	}

	app := &App{
		db: db,
	}

	router := chi.NewRouter()

	router.Use(app.RealIP)
	router.Use(app.RateLimit)
	router.Use(app.LogRequest)

	router.Get("/", app.GetIndex)
	router.Post("/save", app.PostSave())
	router.Get("/{id:[A-Za-z0-9]{8}}", app.GetPaste)
	router.Get("/raw/{id:[A-Za-z0-9]{8}}", app.GetRawPaste)

	router.Get("/*", app.ServeAssets)

	router.NotFound(app.NotFound)
	router.MethodNotAllowed(app.MethodNotAllowed)

	server := &http.Server{
		Addr:         ":80",
		Handler:      router,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println(err)
	}
}
