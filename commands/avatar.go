package commands

import (
	twm "bash06/the-world-machine-v2/bot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func Avatar(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData, b *twm.Bot) {
	user, userOk := data.OptUser("user")
	formatStr, _ := data.OptString("format")

	if !userOk {
		event.CreateMessage(discord.MessageCreate{
			Content: "User option is empty",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	var format discord.FileFormat

	switch formatStr {
	case "jpeg":
		{
			format = discord.FileFormatJPEG
		}
	case "gif":
		{
			format = discord.FileFormatGIF
		}
	default:
		{
			format = discord.FileFormatPNG
		}
	}

	avatarUrl := user.EffectiveAvatarURL(discord.WithFormat(format), discord.WithSize(2048))

	event.CreateMessage(discord.MessageCreate{
		Content: avatarUrl,
	})
}
