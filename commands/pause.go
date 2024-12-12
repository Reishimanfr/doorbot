package commands

import (
	twm "bash06/the-world-machine-v2/bot"
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func Pause(e *events.ApplicationCommandInteractionCreate, data *discord.SlashCommandInteractionData, b *twm.Bot) {
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
			Content: "The music player isn't active. Join a voice channel and `/play` something first!",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	if player.Track() == nil {
		e.CreateMessage(discord.MessageCreate{
			Content: "Nothing is playing right now.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	responseString := ""

	if player.Paused() {
		responseString = "Playback resumed"
		player.Update(context.TODO(), lavalink.WithPaused(false))
	} else {
		responseString = "Playback paused"
		player.Update(context.TODO(), lavalink.WithPaused(true))
	}

	e.CreateMessage(discord.MessageCreate{
		Content: responseString,
		Flags:   discord.MessageFlagEphemeral,
	})
}
