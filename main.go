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
	// Set config
	config, err := config.Load(".")
	if err != nil {
		log.Fatalf("Could not load config. %v", err)
	}
	
	// Set logger
	l := logger.New(config.LogLevel)

	sqldb, err := db.NewSQL("postgres", config.DBConnString, &l)
	if err != nil {
		l.Fatal().Err(err).Send()
	}

	// Do migrations	
	err = migrations.MigrateDB(config.DBConnString, "file://db/migrations/", &l)
	if err != nil {
		l.Fatal().Err(err).Send()
	}

	s := note.NewService(sqldb)

	// Set router
	r, err := server.NewChiRouter(s, config.PASETOSecret, config.AccessTokenDuration, &l)
	if err != nil {
		l.Fatal().Err(err).Send()
	}

	// Set server
	httpServer, err := server.NewHTTP(r, config.HTTPServerAddress, &l)
	if err != nil {
		l.Fatal().Err(err).Send()
	}

	err = httpServer.Shutdown()
	if err != nil {
		l.Fatal().Err(err).Send()
	}

}
