package global

import (
	"flag"

	"github.com/disgoorg/disgo/discord"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
)

var (
	Token = flag.String("token", "", "Your discord bot token")

	LavalinkPassword = flag.String("lavalink-password", "youshallnotpass", "Password to your lavalink server")
	LavalinkPort     = flag.Int("lavalink-port", 2333, "Port on which lavalink is running")
	LavalinkHost     = flag.String("lavalink-host", "localhost", "IP address to your lavalink server")
	LavalinkSecure   = flag.Bool("lavalink-secure", false, "Use a secure connection (HTTPS)")
	LavalinkNodeName = flag.String("lavalink-node-name", "The World Machine v2", "Name for your lavalink node")

	RegisterCommands = flag.Bool("register-commands", true, "Register commands on startup")
	PurgeCommands    = flag.Bool("purge-commands", false, "Purge all (/) commands (quits after they're deleted)")

	UploadAssets = flag.Bool("upload-assets", true, "Upload required assets on startup (or update missing ones)")
	PurgeAssets  = flag.Bool("purge-assets", false, "Purge uploaded assets on startup (quits after they're deleted)")

	EnableLegacyCommands = flag.Bool("enable-legacy-commands", true, "Enables legacy (text) commands alongside (/) commands")

	otelName = "bash06/the-world-machine-v2/self-hosted"
	Tracer   = otel.Tracer(otelName)
	Meter    = otel.Meter(otelName)
	Logger   = otelslog.NewLogger(otelName)

	// Internal variables
	minSeekVal    = 1
	maxSecondsVal = 59

	SlashCommands = []discord.ApplicationCommandCreate{
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
					Description: "New avatar",
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
					Description: "New username",
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
					MinValue:    &minSeekVal,
				},
				discord.ApplicationCommandOptionInt{
					Name:        "seconds",
					Description: "Seconds to seek to",
					Required:    true,
					MinValue:    &minSeekVal,
					MaxValue:    &maxSecondsVal,
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

func ParseFlags() {
	flag.Parse()

	if *Token == "" {
		panic("no bot token set")
	}

	if *LavalinkPassword == "" {
		panic("lavalink password not set")
	}

	if *LavalinkHost == "" {
		panic("lavalink host not set")
	}

	if *LavalinkNodeName == "" {
		panic("lavalink node name not set")
	}

}
