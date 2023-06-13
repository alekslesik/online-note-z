package main

import (
	"log"

	"github.com/alekslesik/online-note-z/db"
	"github.com/alekslesik/online-note-z/db/migrations"
	"github.com/alekslesik/online-note-z/lib/config"
	logger "github.com/alekslesik/online-note-z/lib/logger"
	"github.com/alekslesik/online-note-z/note"
	server "github.com/alekslesik/online-note-z/server/http"
)

func main() {
	// Set cfg
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatalf("Could not load config. %v", err)
	}
	
	// Set logger
	l := logger.New(cfg.LogLevel)

	sqldb, err := db.NewSQL("postgres", cfg.DBConnString, &l)
	if err != nil {
		l.Fatal().Err(err).Send()
	}

	// Do migrations	
	err = migrations.MigrateDB(cfg.DBConnString, "file://db/migrations/", &l)
	if err != nil {
		l.Fatal().Err(err).Send()
	}

	s := note.NewService(sqldb)

	// Set router
	r, err := server.NewChiRouter(s, cfg.PASETOSecret, cfg.AccessTokenDuration, &l)
	if err != nil {
		l.Fatal().Err(err).Send()
	}

	// Set server
	httpServer, err := server.NewHTTP(r, cfg.HTTPServerAddress, &l)
	if err != nil {
		l.Fatal().Err(err).Send()
	}

	err = httpServer.Shutdown()
	if err != nil {
		l.Fatal().Err(err).Send()
	}
}
