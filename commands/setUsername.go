package commands

import (
	twm "bash06/the-world-machine-v2/bot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func SetUsername(event *events.ApplicationCommandInteractionCreate, data *discord.SlashCommandInteractionData, b *twm.Bot) {
	newUsername, ok := data.OptString("username")

	if !ok {
		event.CreateMessage(discord.MessageCreate{
			Content: "No username provided.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	event.Client().Rest().UpdateCurrentUser(discord.UserUpdate{
		Username: newUsername,
		Avatar:   nil,
		Banner:   nil,
	})

	event.CreateMessage(discord.MessageCreate{
		Content: "New username set",
		Flags:   discord.MessageFlagEphemeral,
	})
}
