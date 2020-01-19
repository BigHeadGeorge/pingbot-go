package main

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const token = ""

// waitTime is the time to wait between pings
const waitTime = time.Second * 5

var pings map[string]*ping

func main() {
	pings = make(map[string]*ping)
	d, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	d.AddHandler(messageCreate)

	err = d.Open()
	if err != nil {
		panic(err)
	}
	defer d.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, os.Kill)
	<-sc
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.ToLower(m.Content) == "stop" {
		var found bool
		for userID, ping := range pings {
			if ping.ChannelID == m.ChannelID && (ping.AuthorID == m.Author.ID || userID == m.Author.ID) {
				found = true
				ping.running = false
				delete(pings, userID)
				s.ChannelMessageSend(m.ChannelID, "okay")
				break
			}
		}
		if !found {
			s.ChannelMessageSend(m.ChannelID, "nah")
		}
	} else if strings.HasPrefix(m.Content, "@ping") {
		s.ChannelMessageSend(m.ChannelID, parsePing(s, m))
	}
}

func isNum(c rune) bool {
	return c >= 48 && c <= 57
}

func parsePing(s *discordgo.Session, m *discordgo.MessageCreate) string {
	space := strings.Index(m.Content, " ")
	if space == -1 {
		return "usage: @ping <target>"
	}

	mention := m.Content[space+1:]
	userID := mention[strings.IndexFunc(mention, isNum) : len(mention)-1]
	match := regexp.MustCompile(`<@!?\d+>`).MatchString(mention)
	if !match {
		return "invalid target"
	} else if userID == m.Author.ID {
		return "you can't ping yourself silly"
	} else if userID == s.State.User.ID {
	}

	fmt.Println(mention)
	fmt.Println(userID)
	member, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		return "invalid mention"
	} else if member.User.Bot {
		return "no pinging bots"
	}

	pings[userID] = &ping{ChannelID: m.ChannelID, AuthorID: m.Author.ID}
	go pings[userID].Run(s, mention)

	return "added ping :)"
}

type ping struct {
	ChannelID string
	AuthorID  string
	running   bool
}

func (p *ping) Run(s *discordgo.Session, mention string) {
	p.running = true
	time.Sleep(time.Second)

	for p.running {
		s.ChannelMessageSend(p.ChannelID, mention)
		time.Sleep(waitTime)
	}
}
