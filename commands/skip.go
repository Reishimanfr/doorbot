package commands

import (
	twm "bash06/the-world-machine-v2/bot"
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func Skip(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData, b *twm.Bot) {
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

	queue := b.Queues.Get(event.GuildID().String())
	nextTrack, ok := queue.Next()

	if !ok {
		player.Update(context.Background(), lavalink.WithNullTrack())
	} else {
		player.Update(context.Background(), lavalink.WithTrack(nextTrack))
	}

	event.CreateMessage(discord.MessageCreate{
		Content: "Song skipped.",
		Flags:   discord.MessageFlagEphemeral,
	})

	event.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent("Song skipped.").
		SetEphemeral(true).
		Build(),
	)
}
