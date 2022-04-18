package main

import (
	"fmt"
	"github.com/TomRomeo/jmdict"
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type wotdbot struct {
	JMdict  jmdict.Jmdict
	Discord *discordgo.Session
	DB      *gorm.DB
}

func main() {

	// load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.WithError(err).Error("Error loading .env file")
	}

	// connect to the database using gorm
	var err error
	var db *gorm.DB
	log.Info("Connecting to database...")
	if os.Getenv("BUILD") == "PROD" {
		db, err = gorm.Open(postgres.Open(fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_PASSWORD"))), &gorm.Config{})
		if err != nil {
			panic("failed to connect to postgres database")
		}
	} else {
		db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
		if err != nil {
			panic("failed to connect to sqlite database")
		}
	}
	log.Info("Connected to database!")

	// migrate db
	log.Info("Migrating database...")
	migrateDB(db)
	log.Info("Migrated database!")

	// load JMdict
	log.Info("Loading JMdict...")
	f, err := os.Open("JMdict_e_examp")
	if err != nil {
		log.WithError(err).Fatal("Could not open JMdict")
	}
	log.Info("Loaded JMdict!")

	// parse JMdict
	log.Info("Parsing JMdict...")
	j, _, err := jmdict.LoadJmdict(f)
	if err != nil {
		log.WithError(err).Fatal("Could not parse JMdict")
	}

	// clean up JMdict
	var jmd jmdict.Jmdict
	for _, e := range j.Entries {
		check := false
		for _, s := range e.Sense {
			if len(s.Examples) != 0 && !check {
				check = true
				jmd.Entries = append(jmd.Entries, e)
			}
		}
	}
	log.Info("Parsed JMdict..")

	// Connect to discord
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic("error creating Discord session")
	}

	// open a websocket connection to discord
	err = discord.Open()
	if err != nil {
		panic("error opening connection to Discord")
	}

	log.Info("Connected to Discord")

	wotd := wotdbot{
		JMdict:  jmd,
		Discord: discord,
		DB:      db,
	}

	// register commands
	RegisterCommands(&wotd)

	wotd.sendWOTDMessage()

	// register cron job to send a wotd message every day
	c := cron.New()
	if _, err := c.AddFunc("0 0 * * *", func() {
		wotd.sendWOTDMessage()
	}); err != nil {
		log.WithError(err).Fatal("Could not register cron job")
		return
	}

	c.Run()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Info("Shutting down...")
	if err := discord.Close(); err != nil {
		log.WithError(err).Fatal("Error while shutting down")
	}

}

// function that sends a message to all servers
func (wotd *wotdbot) sendWOTDMessage() {

	log.Info("Sending wotd to all registered guilds")

	// get session
	s := wotd.Discord

	// send wotd to every registered guild
	var guilds []*Guild
	if err := wotd.DB.Find(&guilds).Error; err != nil {
		log.WithError(err).Fatal("Error while getting guilds")
		return
	}
	for _, c := range guilds {

		// get wotd JMdict entry
		wordOfTheDay := wotd.getWOTD()

		examples := make([]string, len(wordOfTheDay.Sense[0].Examples))
		for i, e := range wordOfTheDay.Sense[0].Examples {
			examples[i] = fmt.Sprintf(
				`%s
			|| %s ||
			`, e.Sentences[0].Text, e.Sentences[1].Text)
		}

		var meanings = make([]string, len(wordOfTheDay.Sense))
		for i, s := range wordOfTheDay.Sense {
			meanings[i] = s.Glossary[0].Content
		}

		// set the title
		title := wordOfTheDay.Readings[0].Reading

		if len(wordOfTheDay.Kanji) != 0 {
			title = fmt.Sprintf("%s || %s ||", wordOfTheDay.Kanji[0].Expression, title)
		}

		// create the discord embed
		embed := &discordgo.MessageEmbed{
			Title:       title,
			Description: "",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Meanings",
					Value:  strings.Join(meanings, "\n"),
					Inline: false,
				},
				{
					Name:   "Examples",
					Value:  strings.Join(examples, "\n"),
					Inline: false,
				},
			},
			Color: 0x7289DA,
		}

		content := "Check out the word of the day:"

		// add wotd role if specified
		r, err := wotd.Discord.State.Role(c.GuildID, c.RoleID)
		if err == nil {
			content = fmt.Sprintf("%s %s", r.Mention(), content)
		}

		// send a message with the embeds
		_, _ = s.ChannelMessageSendComplex(c.ChannelID, &discordgo.MessageSend{
			Content: content,
			Embeds:  []*discordgo.MessageEmbed{embed},
		})
	}

}

func (wotd *wotdbot) getWOTD() jmdict.JmdictEntry {

	rand.Seed(time.Now().Unix())
	r := rand.Intn(len(wotd.JMdict.Entries))

	return wotd.JMdict.Entries[r]
}
