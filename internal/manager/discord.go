package manager

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"

	"github.com/keshon/dice-roller/internal/botsdef"
	"github.com/keshon/dice-roller/internal/config"
	"github.com/keshon/dice-roller/internal/db"
)

type GuildManager struct {
	Session       *discordgo.Session
	Bots          map[string]map[string]botsdef.Discord
	commandPrefix string
}

// NewGuildManager creates a new GuildManager with the given discord session and bot instances.
//
// Parameters:
// - session: *discordgo.Session
// - botInstances: map[string]*discord.BotInstance
// Return type: *GuildManager
func NewGuildManager(session *discordgo.Session, bots map[string]map[string]botsdef.Discord) *GuildManager {
	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config:", err)
	}

	return &GuildManager{
		Session:       session,
		Bots:          bots,
		commandPrefix: config.DiscordCommandPrefix,
	}
}

// Start starts the GuildManager.
func (gm *GuildManager) Start() {
	slog.Info("Discord instance of guild manager started")
	gm.Session.AddHandler(gm.Commands)
}

// Commands handles the commands received in a Discord session message.
//
// Parameters:
//   - s: a pointer to the Discord session
//   - m: a pointer to the Discord message received
func (gm *GuildManager) Commands(s *discordgo.Session, m *discordgo.MessageCreate) {
	command, _, err := parseCommand(m.Message.Content, gm.commandPrefix)
	if err != nil {
		slog.Error(err)
		return
	}

	switch command {
	case "about", "v", "help", "h", "roll":
		guildID := m.GuildID
		exists, err := db.DoesGuildExist(guildID)
		if err != nil {
			slog.Errorf("Error checking if guild is registered: %v", err)
			return
		}

		if !exists {
			gm.Session.ChannelMessageSend(m.Message.ChannelID, "Guild must be registered first.\nUse `"+gm.commandPrefix+"register` command.")
			return
		}
	}

	switch command {
	case "register":
		gm.handleRegisterCommand(s, m)
	case "unregister":
		gm.handleUnregisterCommand(s, m)
	}
}

// handleRegisterCommand handles the registration command for the GuildManager.
//
// Parameters:
// - s: The discord session.
// - m: The message create event.
func (gm *GuildManager) handleRegisterCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	channelID := m.Message.ChannelID
	guildID := m.GuildID

	exists, err := db.DoesGuildExist(guildID)
	if err != nil {
		slog.Errorf("Error checking if guild is registered: %v", err)
		return
	}

	if exists {
		gm.Session.ChannelMessageSend(channelID, "Guild is already registered")
		return
	}

	guild := db.Guild{ID: guildID, Name: ""}
	err = db.CreateGuild(guild)
	if err != nil {
		slog.Errorf("Error registering guild: %v", err)
		gm.Session.ChannelMessageSend(channelID, "Error registering guild")
		return
	}

	gm.setupBotInstance(s, guildID)
	gm.Session.ChannelMessageSend(channelID, "Guild registered successfully")
}

// handleUnregisterCommand handles the unregister command for the GuildManager.
//
// Parameters:
// - s: the discordgo Session
// - m: the discordgo MessageCreate
// Return type: none
func (gm *GuildManager) handleUnregisterCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	channelID := m.Message.ChannelID
	guildID := m.GuildID

	exists, err := db.DoesGuildExist(guildID)
	if err != nil {
		slog.Errorf("Error checking if guild is registered: %v", err)
		return
	}

	if !exists {
		gm.Session.ChannelMessageSend(channelID, "Guild is not registered")
		return
	}

	err = db.DeleteGuild(guildID)
	if err != nil {
		slog.Errorf("Error unregistering guild: %v", err)
		gm.Session.ChannelMessageSend(channelID, "Error unregistering guild")
		return
	}

	gm.removeBotInstance(guildID)
	gm.Session.ChannelMessageSend(channelID, "Guild unregistered successfully")
}

// setupBotInstance sets up a bot instance for the given guild.
//
// Parameters:
// - session: pointer to discordgo.Session
// - guildID: string
func (gm *GuildManager) setupBotInstance(session *discordgo.Session, guildID string) {
	id := guildID

	if _, ok := gm.Bots[id]; !ok {
		gm.Bots[id] = make(map[string]botsdef.Discord)
	}

	for _, module := range botsdef.Modules {
		botInstance := botsdef.CreateBotInstance(session, module)
		if botInstance != nil {
			gm.Bots[id][module] = botInstance
			botInstance.Start(id)
		}
	}
}

// removeBotInstance removes a bot instance from the GuildManager's Bots map for the given guildID.
//
// Parameters:
// - guildID string: the ID of the guild from which the bot instance will be removed.
// No return type.
func (gm *GuildManager) removeBotInstance(guildID string) {
	bots, ok := gm.Bots[guildID]
	if !ok {
		return
	}

	// Iterate through modules and remove each bot
	for _, module := range botsdef.Modules {
		if bot, ok := bots[module]; ok {
			bot.Stop()
			delete(bots, module)
		}
	}

	delete(gm.Bots, guildID)
}

// parseCommand parses the input based on the provided pattern
//
// input: the input string to be parsed
// pattern: the pattern to match at the beginning of the input
// string: the parsed command
// string: the parsed parameter
// error: an error if the pattern is not found or no command is found
func parseCommand(input, pattern string) (string, string, error) {
	input = strings.ToLower(input)
	pattern = strings.ToLower(pattern)

	if !strings.HasPrefix(input, pattern) {
		return "", "", nil // fmt.Errorf("pattern not found")
	}

	input = input[len(pattern):]

	words := strings.Fields(input)
	if len(words) == 0 {
		return "", "", fmt.Errorf("no command found")
	}

	command := words[0]
	parameter := ""
	if len(words) > 1 {
		parameter = strings.Join(words[1:], " ")
		parameter = strings.TrimSpace(parameter)
	}
	return command, parameter, nil
}
