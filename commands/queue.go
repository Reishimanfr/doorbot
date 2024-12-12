package commands

import (
	twm "bash06/the-world-machine-v2/bot"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func Queue(e *events.ApplicationCommandInteractionCreate, data *discord.SlashCommandInteractionData, b *twm.Bot) {
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
			Content: "The music player isn't active. Join a voice channel and use `/play` first.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	queue := b.Queues.Get(e.GuildID().String())

	if len(queue.Tracks) <= 0 {
		e.CreateMessage(discord.MessageCreate{
			Content: "The queue is empty. Add something using the `/play` command.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	previousButton := discord.ButtonComponent{
		Label:    "Previous",
		Style:    discord.ButtonStyleSecondary,
		CustomID: "queue-previous-page",
		Emoji: &discord.ComponentEmoji{
			Name: "⬅️",
		},
	}

	nextButton := discord.ButtonComponent{
		Label:    "Next",
		Style:    discord.ButtonStyleSecondary,
		CustomID: "queue-next-page",
		Emoji: &discord.ComponentEmoji{
			Name: "➡️",
		},
	}

	actionRow := discord.NewActionRow(previousButton, nextButton)
	embeds, ok := b.CreateQueueMessage(queue, 6)

	if !ok {
		e.CreateMessage(discord.MessageCreate{
			Content: "Failed to generate queue embeds",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	err := e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{
			embeds[0],
		},
		Components: []discord.ContainerComponent{
			actionRow,
		},
	})

	if err != nil {
		slog.Error("Error while sending queue command message", slog.Any("Error", err))
		return
	}
}
