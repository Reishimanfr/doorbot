package twm

import (
	"bash06/the-world-machine-v2/core"
	"bash06/the-world-machine-v2/database"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"gorm.io/gorm"
)

const (
	idlingIconUrl = "https://images-ext-1.discordapp.net/external/IUsiwXR1vQ0aSvxs5KRTrDYZQs0cdtti0j5mH6_2sHE/%3Fsize%3D96%26quality%3Dlossless/https/cdn.discordapp.com/emojis/1027492467337080872.webp"
	activeIconUrl = "https://media.discordapp.net/attachments/968786035788120099/1134526510334738504/niko.gif"
)

type PlayerMessageInfo struct {
	MessageId snowflake.ID
	ChannelId snowflake.ID
}

type Bot struct {
	Client            bot.Client
	Lavalink          disgolink.Client
	CommandHandlers   map[string]func(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData, bot *Bot)
	ComponentHandlers map[string]func(event *events.ComponentInteractionCreate, data *discord.ComponentInteractionData)
	Queues            *QueueManager
	Db                *gorm.DB
	PlayerMessages    map[string]PlayerMessageInfo
	ApplicationEmojis map[string]string // map[Emoji name]Formatted discord emoji
}

type AdditionalTrackData struct {
	Member *discord.Member
}

func NewBot(db *gorm.DB) *Bot {
	emojisRecord := make([]database.ApplicationEmoji, 0)
	db.Find(&emojisRecord)

	emojis := make(map[string]string, len(emojisRecord))

	for _, v := range emojisRecord {
		emojis[v.Name] = "<:" + v.Name + ":" + v.Id + ">"
	}

	return &Bot{
		Queues: &QueueManager{
			Queues: make(map[string]*Queue),
		},
		PlayerMessages:    make(map[string]PlayerMessageInfo, 0),
		CommandHandlers:   make(map[string]func(event *events.ApplicationCommandInteractionCreate, data discord.SlashCommandInteractionData, b *Bot), 0),
		ComponentHandlers: make(map[string]func(event *events.ComponentInteractionCreate, data *discord.ComponentInteractionData), 0),
		ApplicationEmojis: emojis,
		Db:                db,
	}
}

func (b *Bot) UpdatePlayerMessage(player disgolink.Player, track *lavalink.Track, idling bool) error {
	info := b.PlayerMessages[player.GuildID().String()]
	embed, ok := b.CreateNowPlayingMessage(player, track, idling)

	if !ok {
		return errors.New("failed to generate now playing embed")
	}

	if _, err := b.Client.Rest().GetMessage(info.ChannelId, info.MessageId); err != nil {
		return err
	}

	if _, err := b.Client.Rest().UpdateMessage(info.ChannelId, info.MessageId, discord.MessageUpdate{
		Embeds: &[]discord.Embed{embed},
	}); err != nil {
		return err
	}

	return nil
}

func (b *Bot) CreateQueueMessage(queue *Queue, itemsPerPage int) (embed []discord.Embed, ok bool) {
	if len(queue.Tracks) <= 0 {
		return []discord.Embed{}, false
	}

	embeds := make([]discord.Embed, 0)
	page := 1

	for i := 0; i < itemsPerPage; i += itemsPerPage {
		trackSlice := queue.Tracks[i:itemsPerPage]
		page++

		subEmbedDescription := ""

		for _, v := range trackSlice {
			info := v.Info

			var more AdditionalTrackData

			err := json.Unmarshal(v.UserData, &more)
			if err != nil {
				slog.Error("Failed to unmarshal additional track data", slog.Any("Error", err))
				continue
			}

			subEmbedDescription += fmt.Sprintf("`#%v`: %v - %v (%v)\n* Requested by <@%s>", i, info.Title, info.Author, info.Length.String(), more.Member.User.ID)
		}

		embeds = append(embeds, discord.Embed{
			Author: &discord.EmbedAuthor{
				Name: "There are " + strconv.Itoa(len(queue.Tracks)) + " tracks in the queue",
			},
			Description: subEmbedDescription,
			Footer: &discord.EmbedFooter{
				Text: "Page" + strconv.Itoa(page),
			},
		})
	}

	return embeds, true
}

func (b *Bot) CreateSongProgressBar(player disgolink.Player, full bool) string {
	BEGIN := map[string]string{
		"0.00": b.ApplicationEmojis["b0"],
		"0.01": b.ApplicationEmojis["b10"],
		"0.02": b.ApplicationEmojis["b20"],
		"0.03": b.ApplicationEmojis["b30"],
		"0.04": b.ApplicationEmojis["b40"],
		"0.05": b.ApplicationEmojis["b50"],
		"0.06": b.ApplicationEmojis["b60"],
		"0.07": b.ApplicationEmojis["b70"],
		"0.08": b.ApplicationEmojis["b80"],
		"0.09": b.ApplicationEmojis["b90"],
		"0.1":  b.ApplicationEmojis["b100"],
	}

	CENTER := map[string]string{
		"0":  b.ApplicationEmojis["c0"],
		"1":  b.ApplicationEmojis["c10"],
		"2":  b.ApplicationEmojis["c20"],
		"3":  b.ApplicationEmojis["c30"],
		"4":  b.ApplicationEmojis["c40"],
		"5":  b.ApplicationEmojis["c50"],
		"6":  b.ApplicationEmojis["c60"],
		"7":  b.ApplicationEmojis["c70"],
		"8":  b.ApplicationEmojis["c80"],
		"9":  b.ApplicationEmojis["c90"],
		"10": b.ApplicationEmojis["c100"],
	}

	END := map[string]string{
		"0":  b.ApplicationEmojis["e0"],
		"1":  b.ApplicationEmojis["e10"],
		"2":  b.ApplicationEmojis["e20"],
		"3":  b.ApplicationEmojis["e30"],
		"4":  b.ApplicationEmojis["e40"],
		"5":  b.ApplicationEmojis["e50"],
		"6":  b.ApplicationEmojis["e60"],
		"7":  b.ApplicationEmojis["e70"],
		"8":  b.ApplicationEmojis["e80"],
		"9":  b.ApplicationEmojis["e90"],
		"10": b.ApplicationEmojis["e100"],
	}

	fullProgressBar := fmt.Sprintf("%s%s%s", BEGIN["0.1"], repeatString(CENTER["10"], 8), END["10"])
	emptyProgressBar := fmt.Sprintf("%s%s%s", BEGIN["0.00"], repeatString(CENTER["0"], 8), END["0"])

	position := player.Position()
	track := player.Track()

	if full {
		return fullProgressBar
	}

	if track == nil {
		return emptyProgressBar
	}

	length := track.Info.Length

	f := float64(position.Milliseconds()) / float64(length.Milliseconds())
	songProgress := math.Round(f*100) / 100

	var begin, center, end string

	if songProgress >= 1 {
		return fullProgressBar
	}

	if songProgress <= 0 {
		return emptyProgressBar
	}

	if songProgress <= 0.1 {
		begin = BEGIN[fmt.Sprintf("%.2f", songProgress)]
		center = repeatString(CENTER["0"], 8)
		end = END["0"]
	} else if songProgress > 0.1 && songProgress <= 0.9 {
		begin = BEGIN["0.1"]

		repeat := int(math.Floor(songProgress*10)) % 10
		rest := int(songProgress*100) % 10
		repeatRest := 8 - repeat

		center = repeatString(CENTER["10"], repeat)

		if rest > 0 {
			center += CENTER[fmt.Sprintf("%d", rest)]
		}

		if repeatRest > 0 {
			center += repeatString(CENTER["0"], repeatRest)
		}

		end = END["0"]
	} else {
		begin = BEGIN["0.1"]
		center = repeatString(CENTER["10"], 8)

		rest := int(math.Floor(songProgress*10)) % 10
		end = END[fmt.Sprintf("%d", rest)]
	}

	return begin + center + end
}

func (b *Bot) CreateNowPlayingMessage(player disgolink.Player, track *lavalink.Track, idling bool) (embed discord.Embed, ok bool) {
	empty := discord.Embed{}

	if idling {
		q := b.Queues.Get(player.GuildID().String())

		if q.PreviousTrack == nil {
			return discord.NewEmbedBuilder().
				SetAuthor("Idling...", "", idlingIconUrl).
				SetDescription("`No previous track found...`").
				Build(), true
		}

		previous := q.PreviousTrack
		info := previous.Info

		var more AdditionalTrackData

		err := json.Unmarshal(previous.UserData, &more)

		if err != nil {
			return empty, false
		}

		bar := b.CreateSongProgressBar(player, true)

		embed := discord.NewEmbedBuilder().
			SetAuthor("Idling...", "", idlingIconUrl).
			SetTitle(info.Title).
			SetURL(*info.URI).
			SetDescriptionf("By: **%v**\n\n%v\n%s/%s", info.Author, bar, core.FormatTime(track.Info.Length), core.FormatTime(track.Info.Length)).
			SetThumbnail(*info.ArtworkURL).
			SetColor(9109708).
			SetFooterTextf("Requested by %s", more.Member.User.Username).
			SetFooterIcon(more.Member.User.EffectiveAvatarURL()).
			Build()

		return embed, true
	}

	if track == nil {
		return discord.Embed{}, true
	}

	info := track.Info
	progressBar := b.CreateSongProgressBar(player, false)

	embedColor := 9109708
	title := "Now playing..."
	description := fmt.Sprintf("By **%s**\n\n%v\n%s/%s", info.Author, progressBar, core.FormatTime(player.Position()), core.FormatTime(info.Length))

	if info.IsStream {
		embedColor = 15548997
		title = "🎥🔴 Playing a live stream..."
		description = "By: **" + info.Author + "**"
	}

	if player.Paused() {
		title = "Paused..."
	}

	var more AdditionalTrackData

	err := json.Unmarshal(track.UserData, &more)

	if err != nil {
		return empty, false
	}

	return discord.NewEmbedBuilder().
		SetAuthor(title, *info.URI, activeIconUrl).
		SetTitle(info.Title).
		SetURL(*info.URI).
		SetDescription(description).
		SetThumbnail(*info.ArtworkURL).
		SetColor(embedColor).
		SetFooterTextf("Requested by %s", more.Member.User.Username).
		SetFooterIcon(more.Member.User.EffectiveAvatarURL()).
		Build(), true
}

func repeatString(str string, amount int) string {
	output := ""
	for i := 0; i < amount; i++ {
		output += str
	}
	return output
}

func (b *Bot) CheckIfUserInVc(guildId, userId snowflake.ID) bool {
	_, ok := b.Client.Caches().VoiceState(guildId, userId)

	return ok
}
