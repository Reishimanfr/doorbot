package utils

import (
	"errors"
	"slices"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

func CheckIfDj(c bot.Client, member *discord.Member, djRoleId, guildId snowflake.ID) (bool, error) {
	role, ok := c.Caches().Role(guildId, djRoleId)

	if !ok {
		return false, errors.New("role with this ID doesn't exist")
	}

	return slices.Contains(member.RoleIDs, role.ID), nil
}
