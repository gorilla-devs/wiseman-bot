package commands

import (
	"fmt"
	"log"
	"strconv"
	"wiseman/internal/db"
	"wiseman/internal/entities"
	"wiseman/internal/errors"
	"wiseman/internal/services"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Helpers = append(Helpers, Helper{
		Name:        "setrank",
		Category:    "Administrator Commands",
		Description: "setrank sets the range of levels a role can be assigned to a user",
		Usage:       "setrank <rank_id> <min_xp> <max_xp>",
	})

	services.Commands["setrank"] = SetRank
}

func SetRank(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {

	//check if the user has the required role
	if !services.IsUserAdmin(m.Author.ID, m.ChannelID) {
		return errors.CreateUnauthorizedUserError(m.Author.ID)
	}

	// ctx := context.TODO()
	// rank_id min_xp max_xp
	if len(args) != 3 {
		if len(args) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Not enough arguments")
			return errors.CreateInvalidArgumentError(args[0])
		} else if len(args) > 3 {
			s.ChannelMessageSend(m.ChannelID, "Too many arguments")
			return errors.CreateInvalidArgumentError(args[0])
		}
	}

	rank_id := args[0]
	min_xp, err := strconv.Atoi(args[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error Reading argument, expected an integer number")
		return errors.CreateInvalidArgumentError(args[1])
	}

	max_xp, err := strconv.Atoi(args[2])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error Reading argument, expected an integer number")
		return errors.CreateInvalidArgumentError(args[2])
	}

	if min_xp > max_xp {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Min XP cannot be greater than Max XP"))
		return errors.CreateInvalidArgumentError(args[1] + " must be less than " + args[2])
	}

	customRole := &entities.RoleType{
		Id:       rank_id,
		MinLevel: uint(min_xp),
		MaxLevel: uint(max_xp),
	}

	log.Println("new role created:", customRole)

	err = db.UpdateRoleServer(m.GuildID, *customRole)
	if err != nil {
		return err
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role set from %d to %d level", min_xp, max_xp))

	// now the role exists, what we need to do is add it to the server
	// maybe with s.GuildRoleCreate(m.GuildId) and reallocate all
	// users who match this role to the new role

	return nil
}