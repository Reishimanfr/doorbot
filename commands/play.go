package commands

import (
	twm "bash06/the-world-machine-v2/bot"
	"context"
	"log/slog"
	"regexp"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/json"
)

var (
	urlPattern    = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
	searchPattern = regexp.MustCompile(`^(.{2})search:(.+)`)
)

func Play(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData, b *twm.Bot) {
	voiceState, voiceOk := b.Client.Caches().VoiceState(*event.GuildID(), event.User().ID)

	if !voiceOk {
		event.CreateMessage(discord.MessageCreate{
			Content: "You must be in a voice channel to use this command",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	identifier := data.String("url-or-query")
	source, sourceOk := data.OptString("source")

	if sourceOk {
		identifier = lavalink.SearchType(source).Apply(identifier)
	} else if !urlPattern.MatchString(identifier) && !searchPattern.MatchString(identifier) {
		identifier = lavalink.SearchTypeYouTube.Apply(identifier)
	}

	if err := event.DeferCreateMessage(true); err != nil {
		slog.Error("Error while deferring message", slog.Any("Error", err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	player := b.Lavalink.Player(*event.GuildID())
	queue := b.Queues.Get(event.GuildID().String())
	var tracks []lavalink.Track

	b.Lavalink.BestNode().LoadTracksHandler(ctx, identifier, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			b.Client.Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
				SetContentf("Track `%s - %s` added to the queue", track.Info.Title, track.Info.Author).
				Build(),
			)

			tracks = []lavalink.Track{track}
		},
		// TODO: make this ask the user if they want to load the playlist
		func(playlist lavalink.Playlist) {
			b.Client.Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
				SetContentf("Loaded playlist `%s` with `%d tracks`", playlist.Info.Name, len(playlist.Tracks)).
				Build(),
			)

			tracks = playlist.Tracks

			// TODO: make it so the user can select tracks to add from the playlist using select menus (maybe)
			// queue.Add(playlist.Tracks...)
		},
		func(results []lavalink.Track) {
			b.Client.Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
				SetContentf("Track `%s - %s` added to the queue", results[0].Info.Title, results[0].Info.Author).
				Build(),
			)

			tracks = []lavalink.Track{results[0]}

			// TODO: allow the user to select one of the results
			// queue.Add(tracks[0])
		},
		func() {
			b.Client.Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
				SetContentf("No result found for `%s`", identifier).
				Build(),
			)
		},
		func(err error) {
			b.Client.Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
				SetContentf("Failed to resolve query: `%s`", err.Error()).
				Build(),
			)
		},
	))

	if err := b.Client.UpdateVoiceState(context.TODO(), *event.GuildID(), voiceState.ChannelID, false, true); err != nil {
		b.Client.Rest().CreateFollowupMessage(event.ApplicationID(), event.Token(), discord.NewMessageCreateBuilder().
			SetContentf("Failed to join voice channel: `%s`", err.Error()).
			Build(),
		)
		return
	}

	for _, track := range tracks {
		data, err := json.Marshal(twm.AdditionalTrackData{
			Member: &event.Member().Member,
		})

		if err != nil {
			slog.Error("Failed to marshal additional track data", slog.Any("Error", err))
			continue
		}

		track.UserData = data
		queue.Add(track)
	}

	currentTrack := player.Track()

	if currentTrack == nil && len(queue.Tracks) > 0 {
		nextTrack, ok := queue.Next()

		if ok {
			player.Update(context.Background(), lavalink.WithTrack(nextTrack))
			currentTrack = &nextTrack
		}
	}

	// Now playing message doesn't exist yet
	if _, ok := b.PlayerMessages[event.GuildID().String()]; !ok {
		nowPlayingEmbed, ok := b.CreateNowPlayingMessage(player, currentTrack, false)

		if !ok {
			slog.Error("Now playing embed failed to generate")
			return
		}

		message, err := b.Client.Rest().CreateMessage(event.Channel().ID(), discord.MessageCreate{
			Embeds: []discord.Embed{nowPlayingEmbed},
		})

		if err != nil {
			slog.Error("Error while sending now playing message", slog.Any("ChannelID", event.Channel().ID()), slog.Any("Error", err))
			return
		}

		b.PlayerMessages[event.GuildID().String()] = twm.PlayerMessageInfo{
			MessageId: message.ID,
			ChannelId: message.ChannelID,
		}
	}
}
