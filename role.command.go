package main

import (
	"fmt"
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm/clause"
)

var roleCommand = &discordgo.ApplicationCommandOption{
	Name:        "role",
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Description: "Configure the mentioned wotd role",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionRole,
			Name:        "role",
			Description: "The wotd role",
			Required:    true,
		},
	},
}

func (wotd *wotdbot) roleHandler(i *discordgo.InteractionCreate) {
	if i.Member.Permissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {

		data := i.ApplicationCommandData()

		switch data.Options[0].Name {
		case "role":
			data := data.Options[0]
			choice := data.Options[0].RoleValue(wotd.Discord, i.GuildID)
			if err := wotd.DB.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"role_id"})}).Create(&Guild{
				GuildID: i.GuildID,
				RoleID:  choice.ID,
			}).Error; err != nil {
				// TODO: add error message
				log.Info("test")
			}

			if err := wotd.Discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: uint64(discordgo.MessageFlagsEphemeral),
					Embeds: []*discordgo.MessageEmbed{
						{
							Type:        "rich",
							Title:       "wotd role configured successfully",
							Description: fmt.Sprintf("You have set the wotd role to %s", choice.Mention()),
							Color:       0x00ff00,
							Author: &discordgo.MessageEmbedAuthor{
								Name:    i.Member.User.Username,
								IconURL: i.Member.User.AvatarURL("24px"),
							},
						},
					},
				},
			}); err != nil {
				// TODO: add error message
				log.Info("test")
			}
		}
	}

}
