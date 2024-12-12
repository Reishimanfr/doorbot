package main

import (
	twm "bash06/the-world-machine-v2/bot"
	"bash06/the-world-machine-v2/commands"
	"bash06/the-world-machine-v2/core"
	"bash06/the-world-machine-v2/database"
	"bash06/the-world-machine-v2/global"
	"bash06/the-world-machine-v2/telemetry"
	"context"
	"embed"
	"errors"
	"fmt"
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
	// We use this to embed the assets into the binary so people don't have to
	// put them in a folder somewhere
	//go:embed assets/*
	content embed.FS
)

func main() {
	global.ParseFlags()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	otelShutdown, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		panic(fmt.Errorf("error while initializing opentelemetry: %v", err))
	}

	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	db := database.Init()
	b := twm.New(db)

	client, clientErr := disgo.New(*global.Token,
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

	if *global.RegisterCommands {
		slog.Info("Registering (/) commands...")
		registerCommands(client)
	}

	if *global.PurgeCommands {
		slog.Warn("Removing all application emojis...")
		core.PurgeEmojis(client, db)
	} else {
		core.UploadEmoji(content, client, db)
	}

	b.CommandHandlers = map[string]func(event *events.ApplicationCommandInteractionCreate, data *discord.SlashCommandInteractionData, b *twm.Bot){
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

	b.ComponentHandlers = map[string]func(event *events.ComponentInteractionCreate, data *discord.ComponentInteractionData, b *twm.Bot){}

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
		Address:  *global.LavalinkHost + ":" + strconv.Itoa(*global.LavalinkPort),
		Name:     *global.LavalinkNodeName,
		Password: *global.LavalinkPassword,
		Secure:   *global.LavalinkSecure,
	})

	if lavalinkErr != nil {
		slog.Error("Error while connecting to lavalink", slog.Any("Error", lavalinkErr))
		os.Exit(1)
	}

	slog.Info("Bot up and running!")

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}

func registerCommands(client bot.Client) {
	if _, registerErr := client.Rest().SetGlobalCommands(client.ApplicationID(), global.SlashCommands); registerErr != nil {
		slog.Error("Error while connecting to the gateway", slog.Any("Error", registerErr))
		os.Exit(1)
	}
}
