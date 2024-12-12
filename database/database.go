package database

import (
	"log/slog"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("twm.db"), &gorm.Config{})
	if err != nil {
		slog.Error("Error while connecting to the database", slog.Any("Error", err))
		os.Exit(1)
	}

	db.AutoMigrate(&ApplicationEmoji{})

	return db
}
