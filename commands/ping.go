package commands

import (
	twm "bash06/the-world-machine-v2/bot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func Ping(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData, b *twm.Bot) {
	event.CreateMessage(discord.MessageCreate{
		Content: "Pong! My delay estimate is `" + event.Client().Gateway().Latency().String() + "`",
	})
}
