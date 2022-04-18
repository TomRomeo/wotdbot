package main

import (
	"fmt"
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm/clause"
)

var channelCommand = &discordgo.ApplicationCommandOption{
	Name:        "channel",
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Description: "Configure the channel for wotd",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type: discordgo.ApplicationCommandOptionChannel,
			ChannelTypes: []discordgo.ChannelType{
				discordgo.ChannelTypeGuildText,
			},
			Name:        "channel",
			Description: "The wotd channel",
			Required:    true,
		},
	},
}

func (wotd *wotdbot) channelHandler(i *discordgo.InteractionCreate) {
	if i.Member.Permissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {

		data := i.ApplicationCommandData()

		switch data.Options[0].Name {
		case "channel":
			data := data.Options[0]
			choice := data.Options[0].ChannelValue(wotd.Discord)
			if err := wotd.DB.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"channel_id"})}).Create(&Guild{
				GuildID:   choice.GuildID,
				ChannelID: choice.ID,
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
							Title:       "wotd channel configured successfully",
							Description: fmt.Sprintf("You have set the wotd channel to %s", choice.Mention()),
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
