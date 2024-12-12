package commands

import (
	twm "bash06/the-world-machine-v2/bot"
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func Skip(e *events.ApplicationCommandInteractionCreate, data *discord.SlashCommandInteractionData, b *twm.Bot) {
	memberState, ok := b.Client.Caches().VoiceState(*e.GuildID(), b.Client.ID())
	if !ok {
		e.CreateMessage(discord.MessageCreate{
			Content: "Something went wrong while processing your request. Request ID:",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	if memberState.ChannelID == nil {
		e.CreateMessage(discord.MessageCreate{
			Content: "You must be in a voice channel to use this command.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	botState, ok := b.Client.Caches().VoiceState(*e.GuildID(), b.Client.ID())
	if !ok {
		e.CreateMessage(discord.MessageCreate{
			Content: "Something went wrong while processing your request. Request ID:",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	if memberState.ChannelID != botState.ChannelID && botState.ChannelID != nil {
		e.CreateMessage(discord.MessageCreate{
			Content: "You must be in the same voice channel to use this command.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	player := b.Lavalink.Player(*e.GuildID())

	if !player.State().Connected {
		e.CreateMessage(discord.MessageCreate{
			Content: "Player is inactive. Join a voice channel and use `/play` first.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	queue := b.Queues.Get(e.GuildID().String())
	nextTrack, ok := queue.Next()

	if !ok {
		player.Update(context.Background(), lavalink.WithNullTrack())
	} else {
		player.Update(context.Background(), lavalink.WithTrack(nextTrack))
	}

	e.CreateMessage(discord.MessageCreate{
		Content: "Song skipped.",
		Flags:   discord.MessageFlagEphemeral,
	})

	e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent("Song skipped.").
		SetEphemeral(true).
		Build(),
	)
}
