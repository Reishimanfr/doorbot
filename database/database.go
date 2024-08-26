package database

import (
	"errors"
	"log/slog"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDb() *gorm.DB {
	if _, err := os.Stat("twm.db"); errors.Is(err, os.ErrNotExist) {
		if _, err := os.Create("twm.db"); err != nil {
			slog.Error("Error while creating sqlite database file", slog.Any("Error", err))
			os.Exit(1)
		}
	}

	db, err := gorm.Open(sqlite.Open("twm.db"), &gorm.Config{})
	if err != nil {
		slog.Error("Error while connecting to the database", slog.Any("Error", err))
		os.Exit(1)
	}

	db.AutoMigrate(&ApplicationEmoji{})

	return db
}
