package main

import (
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
	"os"
)

var commands = []*discordgo.ApplicationCommand{

	&discordgo.ApplicationCommand{
		Name:        "wotd",
		Description: "configure wotd bot",
		Options: []*discordgo.ApplicationCommandOption{
			channelCommand,
			roleCommand,
		},
	},
}

func RegisterCommands(wotd *wotdbot) {

	var guildID string
	if os.Getenv("BUILD") != "PROD" {
		guildID = "862337011264126996"
	}
	for _, v := range commands {
		_, err := wotd.Discord.ApplicationCommandCreate(wotd.Discord.State.User.ID, guildID, v)
		if err != nil {
			log.WithError(err).Fatalf("Cannot create '%v' command", v.Name)
		}
	}
	wotd.Discord.AddHandler(wotd.commandHandler)
}

func (wotd *wotdbot) commandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	data := i.ApplicationCommandData()

	switch data.Name {
	case "wotd":
		data := i.ApplicationCommandData()

		switch data.Options[0].Name {
		case "channel":
			wotd.channelHandler(i)
		case "role":
			wotd.roleHandler(i)
		}

	}
}
