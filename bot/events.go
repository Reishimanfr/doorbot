package twm

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func (b *Bot) OnApplicationCommand(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()

	handler, ok := b.CommandHandlers[data.CommandName()]

	if !ok {
		slog.Warn("Unknown command", slog.String("CommandName", data.CommandName()))
		return
	}

	handler(event, &data, b)
}

func (b *Bot) OnVoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	if event.VoiceState.UserID != b.Client.ApplicationID() {
		return
	}

	b.Lavalink.OnVoiceStateUpdate(context.TODO(), event.VoiceState.GuildID, event.VoiceState.ChannelID, event.VoiceState.SessionID)

	if event.VoiceState.ChannelID == nil {
		b.Queues.Delete(event.VoiceState.GuildID.String())
	}

	guildPlayer := b.Lavalink.Player(event.VoiceState.GuildID)

	// Pause if we got guild deafened (or muted)
	if (event.VoiceState.GuildDeaf || event.VoiceState.GuildMute) && !guildPlayer.Paused() {
		guildPlayer.Update(context.TODO(), lavalink.WithPaused(true))
	}

}

func (b *Bot) OnVoiceServerUpdate(event *events.VoiceServerUpdate) {
	b.Lavalink.OnVoiceServerUpdate(context.TODO(), event.GuildID, event.Token, *event.Endpoint)
}

func (b *Bot) OnComponentInteraction(event *events.ComponentInteractionCreate) {
	id := event.Data.CustomID()

	handler, ok := b.ComponentHandlers[id]

	if !ok {
		slog.Warn("No handler for component", slog.String("ComponentId", id))
		return
	}

	handler(event, &event.Data, b)
}
