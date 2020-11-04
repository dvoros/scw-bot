package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type Callback func(s *discordgo.Session, m *discordgo.MessageCreate)

type Commands map[string]Callback

var comm = Commands{
	"ping":       PingCallback,
	"bing":       BingCallback,
	"vpn-start":  VpnStartCallback,
	"vpn-status": VpnStatusCallback,
	"vpn-list":   VpnListCallback,
}
var keywords string

func main() {

	kws := make([]string, len(comm))
	i := 0
	for k, _ := range comm {
		kws[i] = k
		i++
	}
	keywords = strings.Join(kws, ", ")

	token := os.Getenv("SCW_BOT_DISCORD_TOKEN")
	if token == "" {
		log.Fatal("no Discord token provided, please set SCW_BOT_DISCORD_TOKEN environment variable")
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages)
	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	fmt.Println("Bot is now stopping...")

	// Cleanly close down the Discord session.
	dg.Close()
}

func PingCallback(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "Pong!")
}

func BingCallback(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.MessageReactionAdd(m.ChannelID, m.Message.ID, "🚭")
}

func VpnStartCallback(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.MessageReactionAdd(m.ChannelID, m.Message.ID, "🔄")
	s.ChannelMessageSend(m.ChannelID, "🔄 working on it...")

	go func() {
		out, err := exec.Command("/bin/bash", "/root/scw-automation/bin/scw-vpn-start.sh").CombinedOutput()
		if err != nil {
			s.MessageReactionRemove(m.ChannelID, m.Message.ID, "🔄", s.State.User.ID)
			s.MessageReactionAdd(m.ChannelID, m.Message.ID, "❌")
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error starting vpn (%s): %v", strings.TrimSpace(string(out)), err))
			return
		}
		s.MessageReactionRemove(m.ChannelID, m.Message.ID, "🔄", s.State.User.ID)
		s.MessageReactionAdd(m.ChannelID, m.Message.ID, "✅")
		s.ChannelMessageSend(m.ChannelID, "✅ VPN started")
	}()

}

func VpnStatusCallback(s *discordgo.Session, m *discordgo.MessageCreate) {
	out, err := exec.Command("/bin/bash", "/root/scw-automation/bin/scw-vpn-status.sh").CombinedOutput()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ %s", strings.TrimSpace(string(out))))
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ %s", strings.TrimSpace(string(out))))
}

func VpnListCallback(s *discordgo.Session, m *discordgo.MessageCreate) {
	out, err := exec.Command("/bin/bash", "/root/scw-automation/bin/scw-vpn-connection-list.sh").CombinedOutput()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ %s", strings.TrimSpace(string(out))))
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ connected clients:\n%s", strings.TrimSpace(string(out))))
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// debugging
	fmt.Println(m.ChannelID, m.Content)

	// Ignore all messages created by me
	if m.Author.ID == s.State.User.ID {
		return
	}

	mentionedMe := false
	for _, u := range m.Message.Mentions {
		if u.ID == s.State.User.ID {
			mentionedMe = true
			break
		}
	}

	directMessage := m.Message.GuildID == ""

	// Only care about the message if I'm mentioned or sent as DM
	if !mentionedMe && !directMessage {
		return
	}

	for keyword, cb := range comm {
		if strings.Contains(strings.ToLower(m.Message.Content), keyword) {
			cb(s, m)
			return // only execute first match
		}
	}

	// no matching command, reply with help
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(`ℹ️ Nem értem amit mondasz, ezeket mondd: %s`, keywords))
}