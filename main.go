package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"

	"wildber/internal"
	"wildber/internal/config"
)

func main() {
	log.Print("config initializing")
	cfg := config.GetConfig("config/prod.yml")

	log.Print("Creating Application")
	ctx := context.Background()
	db, err := pgxpool.Connect(ctx, cfg.Postgres.Url)
	if err != nil {
		log.Fatal("DB Error:", cfg.Postgres.Url, err)
	}

	log.Print("DB CONNECT")
	defer db.Close()

	app, err := internal.NewApp(cfg, db, ctx)
	if err != nil {
		log.Print(err)
	}

	log.Println("Running Application")
	app.Run()
}
