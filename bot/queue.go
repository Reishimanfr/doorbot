package twm

import (
	"math/rand/v2"

	"github.com/disgoorg/disgolink/v3/lavalink"
)

type QueueType string

type Queue struct {
	Tracks        []lavalink.Track
	Type          QueueType
	PreviousTrack *lavalink.Track
}

type QueueManager struct {
	Queues map[string]*Queue
}

const (
	QueueTypeNormal      QueueType = "normal"
	QueueTypeRepeatTrack QueueType = "repeat_track"
	QueueTypeRepeatQueue QueueType = "repeat_queue"
)

func (q QueueType) String() string {
	switch q {
	case QueueTypeNormal:
		return "Normal"
	case QueueTypeRepeatTrack:
		return "Repeat Track"
	case QueueTypeRepeatQueue:
		return "Repeat Queue"
	default:
		return "unknown"
	}
}

func (q *Queue) Shuffle() {
	rand.Shuffle(len(q.Tracks), func(i, j int) {
		q.Tracks[i], q.Tracks[j] = q.Tracks[j], q.Tracks[i]
	})
}

func (q *Queue) Add(track ...lavalink.Track) {
	q.Tracks = append(q.Tracks, track...)
}

func (q *Queue) Remove(pos int) {
	q.Tracks = append(q.Tracks[:pos], q.Tracks[pos+1:]...)
}

func (q *Queue) Next() (lavalink.Track, bool) {
	if len(q.Tracks) <= 0 {
		return lavalink.Track{}, false
	}

	track := q.Tracks[0]
	q.Tracks = q.Tracks[1:]

	return track, true
}

func (q *Queue) Clear() {
	q.Tracks = make([]lavalink.Track, 0)
}

func (q *QueueManager) Get(guildId string) *Queue {
	queue, ok := q.Queues[guildId]

	if !ok {
		queue = &Queue{
			Tracks: make([]lavalink.Track, 0),
			Type:   QueueTypeNormal,
		}
		q.Queues[guildId] = queue
	}

	return queue
}

func (q *QueueManager) Delete(guildId string) {
	delete(q.Queues, guildId)
}
