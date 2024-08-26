package commands

import (
	twm "bash06/the-world-machine-v2/bot"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func Queue(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData, b *twm.Bot) {
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

	queue := b.Queues.Get(event.GuildID().String())

	if len(queue.Tracks) <= 0 {
		event.CreateMessage(discord.MessageCreate{
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
		event.CreateMessage(discord.MessageCreate{
			Content: "Failed to generate queue embeds",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	err := event.CreateMessage(discord.MessageCreate{
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
