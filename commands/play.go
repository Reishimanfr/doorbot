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

func Play(e *events.ApplicationCommandInteractionCreate, d *discord.SlashCommandInteractionData, b *twm.Bot) {
	memberState, _ := b.Client.Caches().VoiceState(*e.GuildID(), e.User().ID)
	// if !ok {
	// 	e.CreateMessage(discord.MessageCreate{
	// 		Content: "Something went wrong while processing your request. Request ID:",
	// 		Flags:   discord.MessageFlagEphemeral,
	// 	})
	// 	return
	// }

	if memberState.ChannelID == nil {
		e.CreateMessage(discord.MessageCreate{
			Content: "You must be in a voice channel to use this command.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	botState, _ := b.Client.Caches().VoiceState(*e.GuildID(), b.Client.ID())
	// if !ok {
	// 	e.CreateMessage(discord.MessageCreate{
	// 		Content: "Something went wrong while processing your request. Request ID:",
	// 		Flags:   discord.MessageFlagEphemeral,
	// 	})
	// 	return
	// }

	if memberState.ChannelID != botState.ChannelID && botState.ChannelID != nil {
		e.CreateMessage(discord.MessageCreate{
			Content: "You must be in the same voice channel to use this command.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	id := d.String("url-or-query")
	source, ok := d.OptString("source")

	if ok {
		id = lavalink.SearchType(source).Apply(id)
	} else if !urlPattern.MatchString(id) && !searchPattern.MatchString(id) {
		id = lavalink.SearchTypeYouTube.Apply(id)
	}

	if err := e.DeferCreateMessage(true); err != nil {
		slog.Error("Error while deferring message", slog.Any("Error", err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	p := b.Lavalink.Player(*e.GuildID())
	q := b.Queues.Get(e.GuildID().String())
	var tracks []lavalink.Track

	updateReply := b.Client.Rest().UpdateInteractionResponse

	b.Lavalink.BestNode().LoadTracksHandler(ctx, id, disgolink.NewResultHandler(
		// Track loaded
		func(track lavalink.Track) {
			updateReply(e.ApplicationID(), e.Token(), discord.NewMessageUpdateBuilder().
				SetContentf("Track `%s - %s` added to the queue", track.Info.Title, track.Info.Author).
				Build(),
			)

			tracks = []lavalink.Track{track}
		},
		// Playlist loaded
		// TODO: make this ask the user if they want to load the playlist
		func(playlist lavalink.Playlist) {
			updateReply(e.ApplicationID(), e.Token(), discord.NewMessageUpdateBuilder().
				SetContentf("Loaded playlist `%s` with `%d tracks`", playlist.Info.Name, len(playlist.Tracks)).
				Build(),
			)

			tracks = playlist.Tracks
		},
		// Search results
		func(results []lavalink.Track) {
			updateReply(e.ApplicationID(), e.Token(), discord.NewMessageUpdateBuilder().
				SetContentf("Track `%s - %s` added to the queue", results[0].Info.Title, results[0].Info.Author).
				Build(),
			)

			tracks = []lavalink.Track{results[0]}
		},
		// No results
		func() {
			updateReply(e.ApplicationID(), e.Token(), discord.NewMessageUpdateBuilder().
				SetContentf("No result found for `%s`", id).
				Build(),
			)
		},
		// Error
		func(err error) {
			slog.Error("Error while resolving track", slog.String("Track", id), slog.Any("Error", err))
			updateReply(e.ApplicationID(), e.Token(), discord.NewMessageUpdateBuilder().
				SetContentf("Failed to resolve query: `%s`", err.Error()).
				Build(),
			)
		},
	))

	voiceState, ok := b.Client.Caches().VoiceState(*e.GuildID(), e.User().ID)
	if !ok {
		e.CreateMessage(discord.MessageCreate{
			Content: "Failed to get voice state for user",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	if err := b.Client.UpdateVoiceState(context.TODO(), *e.GuildID(), voiceState.ChannelID, false, true); err != nil {
		slog.Error("Error while joining voice channel", slog.Any("Error", err))

		b.Client.Rest().CreateFollowupMessage(e.ApplicationID(), e.Token(), discord.MessageCreate{
			Content: "Something went wrong while joining your voice channel. Please make sure it's not locked",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	for _, track := range tracks {
		data, err := json.Marshal(twm.MoreTrackData{
			Req: twm.TrackRequester{
				Id:        e.User().ID,
				Username:  e.User().Username,
				AvatarURL: e.User().EffectiveAvatarURL(),
			},
		})

		if err != nil {
			slog.Error("Failed to marshal additional track data", slog.Any("Error", err))
			continue
		}

		track.UserData = data
		q.Add(track)
	}

	currentTrack := p.Track()

	if currentTrack == nil && len(q.Tracks) > 0 {
		nextTrack, ok := q.Next()

		if ok {
			p.Update(context.Background(), lavalink.WithTrack(nextTrack))
			currentTrack = &nextTrack
		}
	}

	// Now playing message doesn't exist yet
	if _, ok := b.PlayerMessages[e.GuildID().String()]; !ok {
		nowPlayingEmbed, ok := b.CreatePlayerMessage(*e.GuildID(), false)

		if !ok {
			slog.Error("Now playing embed failed to generate")
			return
		}

		message, err := b.Client.Rest().CreateMessage(e.Channel().ID(), discord.MessageCreate{
			Embeds: []discord.Embed{nowPlayingEmbed},
		})

		if err != nil {
			slog.Error("Error while sending now playing message", slog.Any("ChannelID", e.Channel().ID()), slog.Any("Error", err))
			return
		}

		b.PlayerMessages[e.GuildID().String()] = twm.PlayerMessageInfo{
			MessageId: message.ID,
			ChannelId: message.ChannelID,
		}
	}
}
