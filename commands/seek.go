package commands

import (
	twm "bash06/the-world-machine-v2/bot"
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func Seek(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData, b *twm.Bot) {
	player := b.Lavalink.Player(*event.GuildID())

	if !player.State().Connected {
		event.CreateMessage(discord.MessageCreate{
			Content: "Player is inactive. Join a voice channel and use `/play` first.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	if inVc := b.CheckIfUserInVc(*event.GuildID(), event.User().ID); !inVc {
		event.CreateMessage(discord.MessageCreate{
			Content: "You must be in a voice channel to use this command.",
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

	event.CreateMessage(discord.MessageCreate{
		Content: "Seeked to `" + player.Position().String() + "`.",
		Flags:   discord.MessageFlagEphemeral,
	})
}
