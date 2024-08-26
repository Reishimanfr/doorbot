package main

import (
	twm "bash06/the-world-machine-v2/bot"
	"bash06/the-world-machine-v2/commands"
	"bash06/the-world-machine-v2/core"
	"bash06/the-world-machine-v2/database"
	"context"
	"embed"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/lmittmann/tint"
)

var (
	token       = flag.String("token", "", "discord bot token")
	purgeEmoji  = flag.Bool("purge-emoji", false, "removes all application emojis on start")
	registerCmd = flag.Bool("register-commands", false, "registers (/) commands on startup")

	// Lavalink stuff
	linkPassword = flag.String("link-password", "youshallnotpass", "password to the lavalink server")
	linkPort     = flag.Int("link-port", 2333, "port on which lavalink is running")
	linkHost     = flag.String("link-host", "localhost", "host (IP) on which lavalink is running")
	linkSecure   = flag.Bool("link-secure", false, "should we use a more secure connection?")
	nodeName     = flag.String("node-name", "The World Machine v2", "name given to the node connecting to lavalink")

	//go:embed assets/*
	content embed.FS

	minimumSeekValue = 0
	maxSecondsValue  = 59

	slashCommands = []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "ping",
			Description: "Shows the delay estimate between the bot's host and discord's gateway",
		},
		discord.SlashCommandCreate{
			Name:        "to-gif",
			Description: "Converts an image to a GIF from url or attachment",
			IntegrationTypes: []discord.ApplicationIntegrationType{
				discord.ApplicationIntegrationTypeGuildInstall,
				discord.ApplicationIntegrationTypeUserInstall,
			},
			Contexts: []discord.InteractionContextType{
				discord.InteractionContextTypeBotDM,
				discord.InteractionContextTypeGuild,
				discord.InteractionContextTypePrivateChannel,
			},
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionAttachment{
					Name:        "attachment",
					Description: "Image to convert to gif",
				},
				discord.ApplicationCommandOptionString{
					Name:        "url",
					Description: "URL to the image to be converted to gif",
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "avatar",
			Description: "Returns the avatar of a specified user",
			IntegrationTypes: []discord.ApplicationIntegrationType{
				discord.ApplicationIntegrationTypeGuildInstall,
				discord.ApplicationIntegrationTypeUserInstall,
			},
			Contexts: []discord.InteractionContextType{
				discord.InteractionContextTypeBotDM,
				discord.InteractionContextTypeGuild,
				discord.InteractionContextTypePrivateChannel,
			},
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "User whose avatar to return",
					Required:    true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "format",
					Description: "File format to be used (defaults to PNG)",
					Choices: []discord.ApplicationCommandOptionChoiceString{
						{
							Name:  "PNG",
							Value: "png",
						},
						{
							Name:  "JPEG",
							Value: "jpeg",
						},
						{
							Name:  "GIF",
							Value: "gif",
						},
					},
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "set-avatar",
			Description: "Sets the bot's avatar",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionAttachment{
					Name:        "avatar",
					Description: "Avatar to be set",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "set-username",
			Description: "Sets the bot's username",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "username",
					Description: "New username to be set",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "play",
			Description: "Plays a track from url or search query",
			Options: []discord.ApplicationCommandOption{
				// TODO: implement autocomplete
				discord.ApplicationCommandOptionString{
					Name:        "url-or-query",
					Description: "URL or search query",
					Required:    true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "source",
					Description: "Source to resolve the song from",
					Choices: []discord.ApplicationCommandOptionChoiceString{
						{
							Name:  "Youtube",
							Value: "ytsearch",
						},
						{
							Name:  "Youtube Music",
							Value: "ytmsearch",
						},
						{
							Name:  "Soundcloud",
							Value: "scsearch",
						},
						// TODO: implement more sources
					},
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "pause",
			Description: "Toggles playback of music",
		},
		discord.SlashCommandCreate{
			Name:        "resume",
			Description: "Toggles playback of music",
		},
		discord.SlashCommandCreate{
			Name:        "skip",
			Description: "Skips the currently playing song (and plays the next one if in queue)",
		},
		discord.SlashCommandCreate{
			Name:        "seek",
			Description: "Seeks to a point in the currently playing song",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "minutes",
					Description: "Minute to seek to (tip: this will also seek hours if >59)",
					Required:    true,
					MinValue:    &minimumSeekValue,
				},
				discord.ApplicationCommandOptionInt{
					Name:        "seconds",
					Description: "Seconds to seek to",
					Required:    true,
					MinValue:    &minimumSeekValue,
					MaxValue:    &maxSecondsValue,
				},
				discord.ApplicationCommandOptionString{
					Name:        "mode",
					Description: "Special mode to seek with",
					Choices: []discord.ApplicationCommandOptionChoiceString{
						{
							Name:  "Forward -> seeks by [current time + amount input]",
							Value: "forward",
						},
						{
							Name:  "Backward -> seeks by [current time - amount input]",
							Value: "backward",
						},
					},
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "queue",
			Description: "Shows the current queue",
		},
	}
)

func main() {
	flag.Parse()

	db := database.InitDb()
	b := twm.NewBot(db)

	client, clientErr := disgo.New(*token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentGuildVoiceStates,
			),
		),
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagVoiceStates)),
		bot.WithEventListenerFunc(b.OnApplicationCommand),
		bot.WithEventListenerFunc(b.OnVoiceServerUpdate),
		bot.WithEventListenerFunc(b.OnVoiceStateUpdate),
	)

	if clientErr != nil {
		slog.Error("Error while building disgo instance", slog.Any("Error", clientErr))
		os.Exit(1)
	}

	b.Client = client
	defer client.Close(context.TODO())

	if *registerCmd {
		slog.Info("Registering (/) commands...")
		registerCommands(client)
	}

	if *purgeEmoji {
		slog.Warn("Removing all application emojis...")
		core.PurgeEmojis(client, db)
	} else {
		core.UploadEmojis(content, client, db)
	}

	b.CommandHandlers = map[string]func(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData, b *twm.Bot){
		"avatar":       commands.Avatar,
		"ping":         commands.Ping,
		"set-avatar":   commands.SetAvatar,
		"set-username": commands.SetUsername,
		"play":         commands.Play,
		"pause":        commands.Pause,
		"resume":       commands.Pause,
		"skip":         commands.Skip,
		"seek":         commands.Seek,
		"queue":        commands.Queue,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if gatewayErr := client.OpenGateway(ctx); gatewayErr != nil {
		slog.Error("Error while connecting to gateway", slog.Any("Error", gatewayErr))
		os.Exit(1)
	}

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.Kitchen,
		}),
	))

	b.Lavalink = disgolink.New(client.ApplicationID(),
		disgolink.WithListenerFunc(b.OnPlayerUpdate),
		disgolink.WithListenerFunc(b.OnTrackStart),
		disgolink.WithListenerFunc(b.OnTrackEnd),
		disgolink.WithListenerFunc(b.OnTrackException),
		disgolink.WithListenerFunc(b.OnTrackStuck),
		disgolink.WithListenerFunc(b.OnWebSocketClosed),
		disgolink.WithListenerFunc(b.OnUnknownEvent),
	)

	_, lavalinkErr := b.Lavalink.AddNode(ctx, disgolink.NodeConfig{
		Address:  *linkHost + ":" + strconv.Itoa(*linkPort),
		Name:     *nodeName,
		Password: *linkPassword,
		Secure:   *linkSecure,
	})

	if lavalinkErr != nil {
		slog.Error("Error while connecting to lavalink", slog.Any("Error", lavalinkErr))
		os.Exit(1)
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}

func registerCommands(client bot.Client) {
	if _, registerErr := client.Rest().SetGlobalCommands(client.ApplicationID(), slashCommands); registerErr != nil {
		slog.Error("Error while connecting to the gateway", slog.Any("Error", registerErr))
		os.Exit(1)
	}
}
