package twm

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func (b *Bot) OnPlayerPause(player disgolink.Player, event lavalink.PlayerPauseEvent) {
	err := b.UpdatePlayerMessage(player, player.Track(), false)
	if err != nil {
		slog.Error("Error while updating player now playing message", slog.String("Event", "OnPlayerPause"), slog.Any("Error", err))
		return
	}
}

func (b *Bot) OnPlayerResume(player disgolink.Player, event lavalink.PlayerResumeEvent) {
	err := b.UpdatePlayerMessage(player, player.Track(), false)
	if err != nil {
		slog.Error("Error while updating player now playing message", slog.String("Event", "OnPlayerResume"), slog.Any("Error", err))
		return
	}
}

func (b *Bot) OnPlayerUpdate(player disgolink.Player, event lavalink.PlayerUpdateMessage) {
	// Don't update when we don't have to
	if player.Paused() && player.Track() == nil {
		return
	}

	slog.Info("Updating now playing message")

	err := b.UpdatePlayerMessage(player, player.Track(), false)
	if err != nil {
		slog.Error("Error while updating player now playing message", slog.String("Event", "OnPlayerUpdate"), slog.Any("Error", err))
		return
	}
}

func (b *Bot) OnTrackStart(player disgolink.Player, event lavalink.TrackStartEvent) {
	queue := b.Queues.Get(event.GuildID_.String())
	queue.PreviousTrack = player.Track()
}

func (b *Bot) OnTrackEnd(player disgolink.Player, event lavalink.TrackEndEvent) {
	queue := b.Queues.Get(event.GuildID().String())

	if !event.Reason.MayStartNext() {
		if err := b.UpdatePlayerMessage(player, queue.PreviousTrack, true); err != nil {
			slog.Error("Error while updating player now playing message", slog.String("Event", "OnTrackStart"), slog.Any("Error", err))
		}
		return
	}

	var nextTrack lavalink.Track
	var ok bool

	switch queue.Type {
	case QueueTypeNormal:
		nextTrack, ok = queue.Next()

	case QueueTypeRepeatTrack:
		nextTrack = event.Track

	case QueueTypeRepeatQueue:
		queue.Add(event.Track)
		nextTrack, ok = queue.Next()
	}

	if !ok {
		return
	}

	if err := player.Update(context.TODO(), lavalink.WithTrack(nextTrack)); err != nil {
		slog.Error("Failed to play next track", slog.Any("err", err))
	}
}

func (b *Bot) OnTrackException(player disgolink.Player, event lavalink.TrackExceptionEvent) {
	slog.Info("track exception", slog.Any("event", event))
}

func (b *Bot) OnTrackStuck(player disgolink.Player, event lavalink.TrackStuckEvent) {
	slog.Info("track stuck", slog.Any("event", event))
}

func (b *Bot) OnWebSocketClosed(player disgolink.Player, event lavalink.WebSocketClosedEvent) {
	slog.Info("websocket closed", slog.Any("event", event))
}

func (b *Bot) OnUnknownEvent(p disgolink.Player, e lavalink.UnknownEvent) {
	slog.Info("unknown event", slog.Any("event", e.Type()), slog.String("data", string(e.Data)))
}
