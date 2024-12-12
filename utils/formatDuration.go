package utils

import (
	"fmt"

	"github.com/disgoorg/disgolink/v3/lavalink"
)

// Formats durations into HH?:MM:SS
func FormatTime(duration lavalink.Duration) string {
	hours := duration.HoursPart()
	minutes := duration.MinutesPart()
	seconds := duration.SecondsPart()

	hString := fmt.Sprintf("%d", hours)
	mString := fmt.Sprintf("%d", minutes)
	sString := fmt.Sprintf("%d", seconds)

	if hours < 10 {
		hString = fmt.Sprintf("0%d", hours)
	}

	if minutes < 10 {
		mString = fmt.Sprintf("0%d", minutes)
	}

	if seconds < 10 {
		sString = fmt.Sprintf("0%d", seconds)
	}

	if hours <= 0 {
		return fmt.Sprintf("%s:%s", mString, sString)
	}

	return fmt.Sprintf("%s:%s:%s", hString, mString, sString)
}
