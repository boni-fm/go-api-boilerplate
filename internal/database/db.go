package database

import (
	"context"

	"github.com/boni-fm/go-libsd3/pkg/db/postgres"
	"github.com/boni-fm/go-libsd3/pkg/log"
)

var Db *postgres.Database

func InitDatabase(kunci string, log *log.Logger) {
	dbcfg := postgres.Config{
		KodeDC: kunci,
	}
	db, err := postgres.NewDatabase(context.Background(), &dbcfg)
	if err != nil {
		log.Panicf("Failed to connect to database: %v", err)
		panic(err)
	}
	Db = db
}

func GetDatabase() *postgres.Database {
	return Db
}
