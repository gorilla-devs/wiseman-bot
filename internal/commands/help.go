package commands

import (
	"fmt"
	"log"
	"strings"
	"time"
	"wiseman/internal/services"

	"github.com/bwmarrin/discordgo"
)

type Helper struct {
	Name        string
	Category    string
	Description string
	Usage       string
}

var Helpers []Helper

func init() {
	services.Commands["help"] = Help
}

func Help(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {

	if len(args) == 0 {

		fields := make([]*discordgo.MessageEmbedField, len(Helpers))
		for i, v := range Helpers {
			h := discordgo.MessageEmbedField{
				Name:   v.Name,
				Value:  v.Description,
				Inline: false,
			}
			fields[i] = &h
		}
		fmt.Println(fields)

		embed := &discordgo.MessageEmbed{
			Author:    &discordgo.MessageEmbedAuthor{},
			Color:     9004799,
			Fields:    fields,
			Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
			Title:     "Help",
		}

		_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
		if err != nil {
			log.Println(err)
		}

	} else {

		arg := strings.ToLower(args[0])

		var found bool
		var field []*discordgo.MessageEmbedField

		for _, v := range Helpers {
			if strings.ToLower(v.Name) == arg {
				found = true
				field = append(field, &discordgo.MessageEmbedField{
					Name:   v.Name,
					Value:  v.Description + "\n" + v.Usage,
					Inline: false,
				})
			}
		}

		if !found {
			_, err := s.ChannelMessageSend(m.ChannelID, "No help found for `"+args[0]+"` :frowning:")
			if err != nil {
				log.Println(err)
			}
			return nil
		}

		embed := &discordgo.MessageEmbed{
			Author:    &discordgo.MessageEmbedAuthor{},
			Color:     9004799,
			Fields:    field,
			Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
			Title:     "Help",
		}

		_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}
