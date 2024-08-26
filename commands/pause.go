package commands

import (
	twm "bash06/the-world-machine-v2/bot"
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func Pause(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData, b *twm.Bot) {
	ok := b.CheckIfUserInVc(*event.GuildID(), event.User().ID)

	if !ok {
		event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("You must be in a voice channel to use this command.").
			SetEphemeral(true).
			Build(),
		)
		return
	}

	player := b.Lavalink.Player(*event.GuildID())

	if !player.State().Connected {
		event.CreateMessage(discord.MessageCreate{
			Content: "Player is inactive. Join a voice channel and use `/play` first.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	if player.Track() == nil {
		event.CreateMessage(discord.MessageCreate{
			Content: "Nothing is playing right now.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	responseString := ""

	if player.Paused() {
		responseString = "Resumed playback"
		player.Update(context.TODO(), lavalink.WithPaused(false))
	} else {
		responseString = "Paused playback"
		player.Update(context.TODO(), lavalink.WithPaused(true))
	}

	event.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent(responseString).
		SetEphemeral(true).
		Build(),
	)
}
