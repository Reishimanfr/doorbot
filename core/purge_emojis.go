package core

import (
	"bash06/the-world-machine-v2/database"
	"log/slog"
	"os"

	"github.com/disgoorg/disgo/bot"
	"gorm.io/gorm"
)

func PurgeEmojis(client bot.Client, db *gorm.DB) {
	emojiData, err := client.Rest().GetApplicationEmojis(client.ApplicationID())
	if err != nil {
		slog.Error("Error while getting application emojis", slog.Any("Error", err))
		os.Exit(1)
	}

	for _, v := range emojiData {
		err := client.Rest().DeleteApplicationEmoji(client.ApplicationID(), v.ID)
		if err != nil {
			slog.Error("Error while deleting application emoji", slog.String("EmojiName", v.Name), slog.Any("Error", err))
			os.Exit(1)
		}

		slog.Warn("Deleted application emoji", slog.String("Emoji", v.String()))
	}

	db.Unscoped().Where("1 = 1").Delete(&database.ApplicationEmoji{})
	slog.Info("Deleted all database records")
}
