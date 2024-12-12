package twm

import (
	"bash06/the-world-machine-v2/database"
	"bash06/the-world-machine-v2/utils"
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
	"github.com/disgoorg/snowflake/v2"
	"gorm.io/gorm"
)

type Const struct {
}

const (
	purpleInt        = 9109708
	redInt           = 15548997
	playingString    = "Now playing..."
	liveStreamString = "🎥🔴 Playing a live stream..."
	pausedString     = "Paused..."
	// TODO: this should just be an emoji
	idlingIconURL = "https://images-ext-1.discordapp.net/external/IUsiwXR1vQ0aSvxs5KRTrDYZQs0cdtti0j5mH6_2sHE/%3Fsize%3D96%26quality%3Dlossless/https/cdn.discordapp.com/emojis/1027492467337080872.webp"
	activeIconURL = "https://media.discordapp.net/attachments/968786035788120099/1134526510334738504/niko.gif"
)

var (
	emptyEmbed = discord.Embed{}
)

type PlayerMessageInfo struct {
	MessageId snowflake.ID
	ChannelId snowflake.ID
}

type Bot struct {
	Client            bot.Client
	Lavalink          disgolink.Client
	CommandHandlers   map[string]func(event *events.ApplicationCommandInteractionCreate, data *discord.SlashCommandInteractionData, bot *Bot)
	ComponentHandlers map[string]func(event *events.ComponentInteractionCreate, data *discord.ComponentInteractionData, bot *Bot)
	Queues            *QueueManager
	Db                *gorm.DB
	PlayerMessages    map[string]PlayerMessageInfo
	ApplicationEmojis map[string]string // map[Emoji name]Formatted discord emoji
}

type TrackRequester struct {
	Id        snowflake.ID
	Username  string
	AvatarURL string
}

type MoreTrackData struct {
	Req TrackRequester
}

func New(db *gorm.DB) *Bot {
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
		CommandHandlers:   make(map[string]func(event *events.ApplicationCommandInteractionCreate, data *discord.SlashCommandInteractionData, b *Bot), 0),
		ComponentHandlers: make(map[string]func(event *events.ComponentInteractionCreate, data *discord.ComponentInteractionData, b *Bot), 0),
		ApplicationEmojis: emojis,
		Db:                db,
	}
}

// Music player methods
func (b *Bot) CreatePlayerMessage(guildId snowflake.ID, idling bool) (discord.Embed, bool) {
	player := b.Lavalink.Player(guildId)
	queue := b.Queues.Get(guildId.String())
	track := player.Track()

	if idling {
		if queue.PreviousTrack == nil {
			return discord.Embed{
				Author: &discord.EmbedAuthor{
					Name:    "Idling...",
					IconURL: idlingIconURL,
					URL:     "",
				},
				Description: "`No previous track found...`",
			}, true
		}

		prev := queue.PreviousTrack
		info := prev.Info

		var more MoreTrackData

		err := json.Unmarshal(prev.UserData, &more)
		if err != nil {
			return emptyEmbed, false
		}

		bar := b.CreateProgressBar(player, true)

		return discord.Embed{
			Author: &discord.EmbedAuthor{
				Name:    "Idling...",
				IconURL: idlingIconURL,
			},
			Title:       info.Title,
			URL:         *info.URI,
			Description: fmt.Sprintf("By: **%v**\n\n%v\n%s/%s", info.Author, bar, utils.FormatTime(info.Length), utils.FormatTime(info.Length)),
			Footer: &discord.EmbedFooter{
				Text:    "Requested by " + more.Req.Username,
				IconURL: more.Req.AvatarURL,
			},
		}, true
	}

	if track == nil {
		return emptyEmbed, true
	}

	info := track.Info

	bar := b.CreateProgressBar(player, false)

	embedColor := purpleInt
	title := playingString
	description := fmt.Sprintf("By **%s**\n\n%v\n%s/%s", info.Author, bar, utils.FormatTime(player.Position()), utils.FormatTime(info.Length))

	if info.IsStream {
		embedColor = redInt
		title = liveStreamString
		description = "By: **" + info.Author + "**"
	}

	if player.Paused() {
		title = "Paused..."
	}

	var more MoreTrackData

	err := json.Unmarshal(track.UserData, &more)
	if err != nil {
		return emptyEmbed, false
	}

	return discord.Embed{
		Author: &discord.EmbedAuthor{
			Name:    title,
			URL:     *info.URI,
			IconURL: activeIconURL,
		},
		Title:       info.Title,
		URL:         *info.URI,
		Description: description,
		Thumbnail: &discord.EmbedResource{
			URL: *info.ArtworkURL,
		},
		Color: embedColor,
		Footer: &discord.EmbedFooter{
			Text:    "Requested by " + more.Req.Username,
			IconURL: more.Req.AvatarURL,
		},
	}, true
}

func (b *Bot) UpdatePlayerMessage(guildId snowflake.ID, idling bool) error {
	m := b.PlayerMessages[guildId.String()]
	embed, ok := b.CreatePlayerMessage(guildId, idling)

	if !ok {
		return errors.New("failed to generate now playing embed")
	}

	_, err := b.Client.Rest().UpdateMessage(m.ChannelId, m.MessageId, discord.MessageUpdate{
		Embeds: &[]discord.Embed{embed},
	})

	return err
}

// Future me: please forgive me for this terrible code
// only the current me and god know how this works
// don't expect to optimize this, you will most likely fail
//
// wasted_hours = 4
func (b *Bot) CreateProgressBar(player disgolink.Player, full bool) string {
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

	fullBar := BEGIN["0.1"] + utils.RepeatString(CENTER["10"], 8) + END["10"]
	emptyBar := BEGIN["0.00"] + utils.RepeatString(CENTER["0"], 8) + END["0"]

	position := player.Position()
	track := player.Track()

	if full {
		return fullBar
	}

	if track == nil {
		return emptyBar
	}

	length := track.Info.Length

	f := float64(position.Milliseconds()) / float64(length.Milliseconds())
	songProgress := math.Round(f*100) / 100

	var begin, center, end string

	if songProgress >= 1 {
		return fullBar
	}

	if songProgress <= 0 {
		return emptyBar
	}

	if songProgress <= 0.1 {
		begin = BEGIN[fmt.Sprintf("%.2f", songProgress)]
		center = utils.RepeatString(CENTER["0"], 8)
		end = END["0"]
	} else if songProgress > 0.1 && songProgress <= 0.9 {
		begin = BEGIN["0.1"]

		repeat := int(math.Floor(songProgress*10)) % 10
		rest := int(songProgress*100) % 10
		repeatRest := 8 - repeat

		center = utils.RepeatString(CENTER["10"], repeat)

		if rest > 0 {
			center += CENTER[fmt.Sprintf("%d", rest)]
		}

		if repeatRest > 0 {
			center += utils.RepeatString(CENTER["0"], repeatRest)
		}

		end = END["0"]
	} else {
		begin = BEGIN["0.1"]
		center = utils.RepeatString(CENTER["10"], 8)

		rest := int(math.Floor(songProgress*10)) % 10
		end = END[fmt.Sprintf("%d", rest)]
	}

	return begin + center + end
}

func (b Bot) CreateQueueMessage(queue *Queue, itemsPerPage int) (pages []discord.Embed, ok bool) {
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

			var more MoreTrackData

			err := json.Unmarshal(v.UserData, &more)
			if err != nil {
				slog.Error("Failed to unmarshal additional track data", slog.Any("Error", err))
				continue
			}

			subEmbedDescription += fmt.Sprintf("`#%v`: %v - %v (%v)\n* Requested by <@%s>", i, info.Title, info.Author, info.Length.String(), more.Req.Id)
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
