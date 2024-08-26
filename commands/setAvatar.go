package commands

import (
	twm "bash06/the-world-machine-v2/bot"
	"net/http"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/json"
)

func SetAvatar(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData, b *twm.Bot) {
	avatarFile, ok := data.OptAttachment("avatar")

	if !ok {
		event.CreateMessage(discord.MessageCreate{
			Content: "Avatar attachment is missing.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	resp, err := http.Get(avatarFile.URL)

	if err != nil {
		event.CreateMessage(discord.MessageCreate{
			Content: "Failed to get the avatar file.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	defer resp.Body.Close()

	avatarIcon, err := discord.NewIcon(discord.IconTypeJPEG, resp.Body)

	if err != nil {
		event.CreateMessage(discord.MessageCreate{
			Content: "Failed to create icon",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	event.Client().Rest().UpdateCurrentUser(discord.UserUpdate{
		Username: "",
		Avatar:   json.NewNullablePtr(*avatarIcon),
		Banner:   nil,
	})

	event.CreateMessage(discord.MessageCreate{
		Content: "New profile picture set!",
		Flags:   discord.MessageFlagEphemeral,
	})
}
