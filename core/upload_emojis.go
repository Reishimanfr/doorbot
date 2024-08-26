package core

import (
	"bash06/the-world-machine-v2/database"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"gorm.io/gorm"
)

func UploadEmojis(content embed.FS, client bot.Client, db *gorm.DB) {
	emojiFiles, err := content.ReadDir("assets")
	if err != nil {
		slog.Error("Error reading assets directory", slog.Any("Error", err))
		os.Exit(1)
	}

	currentPath, err := os.Getwd()
	if err != nil {
		slog.Error("Error getting current directory", slog.Any("Error", err))
		os.Exit(1)
	}

	existingEmojiData, err := client.Rest().GetApplicationEmojis(client.ApplicationID())
	if err != nil {
		slog.Error("Error getting client emojis", slog.Any("Error", err))
		os.Exit(1)
	}

	var emojiRecords []database.ApplicationEmoji
	db.Find(&emojiRecords)

	existingEmojis := []string{}
	emojiRecordNames := []string{}

	for _, v := range existingEmojiData {
		existingEmojis = append(existingEmojis, v.Name)
	}

	for _, v := range emojiRecords {
		emojiRecordNames = append(emojiRecordNames, v.Name)
	}

	for _, file := range emojiFiles {
		nameWithoutExt := strings.Split(file.Name(), ".")[0]

		if slices.Contains(existingEmojis, nameWithoutExt) {
			continue
		}

		fullPath := path.Join(currentPath, "./assets/"+file.Name())

		f, err := os.Open(fullPath)
		if err != nil {
			slog.Error("Error while opening file", slog.Any("Error", err))
			os.Exit(1)
		}

		icon, err := discord.NewIcon(discord.IconTypePNG, f)
		if err != nil {
			slog.Error("Error while creating new icon", slog.Any("Error", err))
			os.Exit(1)
		}

		emoji, err := client.Rest().CreateApplicationEmoji(client.ApplicationID(), discord.EmojiCreate{
			Name:  nameWithoutExt,
			Image: *icon,
		})
		if err != nil {
			slog.Error("Error while creating application emoji", slog.Any("Error", err))
			os.Exit(1)
		}

		if !slices.Contains(emojiRecordNames, nameWithoutExt) {
			db.Create(database.ApplicationEmoji{
				Name:     nameWithoutExt,
				Id:       emoji.ID.String(),
				Animated: emoji.Animated,
			})
		}

		fmt.Printf("\nCreated application emoji with name %s", nameWithoutExt)
	}

}
