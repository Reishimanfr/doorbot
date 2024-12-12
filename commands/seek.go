package commands

import (
	twm "bash06/the-world-machine-v2/bot"
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func Seek(e *events.ApplicationCommandInteractionCreate, data *discord.SlashCommandInteractionData, b *twm.Bot) {
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

	if player.Track() == nil {
		e.CreateMessage(discord.MessageCreate{
			Content: "Nothing is playing right now.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	minutes, _ := data.OptInt("minutes")
	seconds, _ := data.OptInt("seconds")
	modifier, _ := data.OptString("mode")

	timeSeconds := int((minutes * 60) + seconds)

	if modifier == "forward" {
		timeSeconds += int(player.Position().Seconds())
	} else if modifier == "backward" {
		timeSeconds = int(player.Position().Seconds()) - timeSeconds

		if timeSeconds < 0 {
			timeSeconds = 0
		}
	}

	player.Update(context.Background(), lavalink.WithPosition(lavalink.Second*lavalink.Duration(timeSeconds)))

	e.CreateMessage(discord.MessageCreate{
		Content: "Seeked to `" + player.Position().String() + "`.",
		Flags:   discord.MessageFlagEphemeral,
	})
}
