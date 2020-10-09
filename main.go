package main

import (
	"flag"
	"fmt"
	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"strings"
)

var (
	commandPrefix string
	botID         string
	debug         bool
)

var Session, _ = discordgo.New()
var zero = int64(0)

func init() {
	Session.Token = os.Getenv("DG_TOKEN")
	if Session.Token == "" {
		flag.StringVar(&Session.Token, "t", "", "Discord Auth Token")
	}

	commandPrefix = os.Getenv("DG_COMMAND_PREFIX")
	if commandPrefix == "" {
		flag.StringVar(&commandPrefix, "cp", "!", "Discord command prefix")
	}

	flag.BoolVar(&debug, "debug", false, "Enable debug message logger mode")
}

func main() {
	flag.Parse()

	if Session.Token == "" {
		log.Fatal("A discord token must be provided")
		return
	}

	discord, err := discordgo.New("Bot " + Session.Token)
	errCheck("error creating discord session", err)

	// make sure we have a user account
	user, err := discord.User("@me")
	errCheck("error retrieving account", err)
	log.Printf("Running as %s\n", user.Username)
	log.Printf("Command prefix is %s\n", commandPrefix)

	discord.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		err = discord.UpdateStatus(0, "Staring daggers at the other bots")
		if err != nil {
			fmt.Println("Error attempting to set my status")
		}
		guilds := discord.State.Guilds
		log.Printf("Started on %d servers", len(guilds))
		for _, guild := range guilds {
			// TODO need to figure out how to get the guild names, probably needs more permissions (?)
			log.Printf("\t%s - %s\n", guild.ID, guild.Name)
		}
	})
	// create the router
	router := exrouter.New()

	router.On("hostile", func(ctx *exrouter.Context) {
		log.Print(ctx.Msg.Content)
        content := strings.Split(ctx.Msg.Content, " ")
        if len(content) > 1 {
            channelName := strings.Join(content[1:], " ")
            channels, err := discord.GuildChannels(ctx.Msg.GuildID)
            if nil != err {
                log.Print("error reading channels: " + err.Error())
                return
            }
    		for _, channel := range channels {
                log.Printf("looking at channel %s", channel.Name)
                if (channelName == channel.Name ) {
                    if (channel.ID != ctx.Msg.ChannelID) {
                        messages, err := discord.ChannelMessages(channel.ID, 100, "", "", "")

                        if nil != err {
                            log.Print("error reading messages: " + err.Error())
                            return
                        }

                        for i := len(messages) - 1; i >= 0; i-- {
                            if !strings.HasPrefix(messages[i].Content, commandPrefix+"roe") {
                                log.Printf("%s\n%s\n%s\n\n", messages[i].ID, messages[i].Content, messages[i].Timestamp)
                                ctx.Reply(messages[i].Content)
                            }
                        }
                        return
                    } else {
                        ctx.Reply("You can't use this in the channel you're in")
                        return
                    }
                }
            }
    		ctx.Reply("No channel defined for " + channelName)
        }
        ctx.Reply("You'll need to supply the name of a family of hostiles")
	})

	// add the router as a handler
	discord.AddHandler(messageLogger)
	discord.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		router.FindAndExecute(discord, commandPrefix, discord.State.User.ID, m.Message)
	})

	// connect to discord
	err = discord.Open()
	errCheck("Error opening connection to Discord", err)

	log.Println("Bot is now running")
	<-make(chan struct{})
}

func errCheck(msg string, err error) {
	if err != nil {
		log.Fatalf("%s %s\n", msg, err)
		panic(err)
	}
}

func messageLogger(session *discordgo.Session, message *discordgo.MessageCreate) {
	if debug {
		// no need to log our own messages
		if session.State.User.ID == message.Author.ID {
			return
		}

		log.Printf("%s %s %s %s\n", message.GuildID, message.ChannelID, message.Author.Username, message.Content)
	}
}

func limit(mesg string) bool {
	return len(mesg) < 1990
}

func chunkMessage(ctx *exrouter.Context, header string, payload string) {
	i := 0
	mesg := ""
	parts := strings.Split(payload, "\n")

	for i < len(parts) {
		if limit(header + mesg + parts[i] + "\n") {
			mesg += parts[i] + "\n"
		} else {
			_, err := ctx.Reply(header + mesg)
			if err != nil {
				log.Printf("Error sending grudges list '%s'", err.Error())
			}
			mesg = ""
		}
		i++
	}

	if len(mesg) > 0 {
		_, err := ctx.Reply(header + mesg)
		if err != nil {
			log.Printf("Error sending grudges list '%s'", err.Error())
		}
	}
}
