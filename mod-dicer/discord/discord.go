package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/dice-roller/internal/config"
	"github.com/keshon/dice-roller/mod-dicer/utils"
)

// BotInstance represents an instance of a Discord bot.
type BotInstance struct {
	DiceRoller *Discord
}

// Discord represents the Melodix instance for Discord.
type Discord struct {
	Session              *discordgo.Session
	GuildID              string
	IsInstanceActive     bool
	prefix               string
	lastChangeAvatarTime time.Time
	rateLimitDuration    time.Duration
}

// NewDiscord creates a new instance of Discord.
func NewDiscord(session *discordgo.Session) *Discord {
	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	return &Discord{
		Session:           session,
		IsInstanceActive:  true,
		prefix:            config.DiscordCommandPrefix,
		rateLimitDuration: time.Minute * 10,
	}
}

// Start starts the Discord instance.
func (d *Discord) Start(guildID string) {
	slog.Infof(`Discord instance started for guild id %v`, guildID)

	d.Session.AddHandler(d.Commands)
	d.GuildID = guildID
}

func (d *Discord) Stop() {
	d.IsInstanceActive = false
}

// Commands handles incoming Discord commands.
func (d *Discord) Commands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != d.GuildID {
		return
	}

	if !d.IsInstanceActive {
		return
	}

	slog.Info("User input: ", m.Message.Content)

	command, parameter, err := parseCommand(m.Message.Content, d.prefix)
	if err != nil {
		return
	}

	slog.Warnf("Command: %v, Parameter: %v", command, parameter)

	commandAliases := [][]string{
		{"roll", "r"},
	}

	canonicalCommand := getCanonicalCommand(command, commandAliases)
	if canonicalCommand == "" {
		return
	}

	slog.Warn("Canonical command is ", canonicalCommand)
	slog.Warn("Len of Canonical command is ", len(canonicalCommand))

	switch canonicalCommand {
	case "roll":
		d.handleRollCommand(s, m, parameter)

	default:
		// Unknown command
	}
}

// parseCommand parses the command and parameter from the Discord input based on the provided pattern.
func parseCommand(content, pattern string) (string, string, error) {
	// Convert both content and pattern to lowercase for case-insensitive comparison
	content = strings.ToLower(content)
	pattern = strings.ToLower(pattern)

	if !strings.HasPrefix(content, pattern) {
		return "", "", fmt.Errorf("Pattern not found")
	}

	content = content[len(pattern):] // Strip the pattern

	words := strings.Fields(content) // Split by whitespace, handling multiple spaces
	if len(words) == 0 {
		return "", "", fmt.Errorf("No command found")
	}

	command := words[0]
	parameter := ""
	if len(words) > 1 {
		parameter = strings.Join(words[1:], " ")
		parameter = strings.TrimSpace(parameter)
	}
	return command, parameter, nil
}

func parseCommandAndParameter(content, pattern string) (string, string, error) {
	if !strings.HasPrefix(content, pattern) {
		return "", "", fmt.Errorf("pattern not found")
	}

	content = content[len(pattern):]

	words := strings.Fields(content)
	if len(words) == 0 {
		return "", "", fmt.Errorf("no command found")
	}

	command := strings.ToLower(words[0])
	parameter := ""
	if len(words) > 1 {
		parameter = strings.Join(words[1:], " ")
		parameter = strings.TrimSpace(parameter)
	}
	return command, parameter, nil
}

func getCanonicalCommand(alias string, commandAliases [][]string) string {
	for _, aliases := range commandAliases {
		for _, a := range aliases {
			if a == alias {
				return aliases[0]
			}
		}
	}
	return ""
}

func (d *Discord) changeAvatar(s *discordgo.Session) {
	if time.Since(d.lastChangeAvatarTime) < d.rateLimitDuration {
		//slog.Info("Rate-limited. Skipping changeAvatar.")
		return
	}

	imgPath, err := utils.GetWeightedRandomImagePath("./assets/avatars")
	if err != nil {
		slog.Errorf("Error getting avatar path: %v", err)
		return
	}

	avatar, err := utils.ReadFileToBase64(imgPath)
	if err != nil {
		fmt.Printf("Error preparing avatar: %v\n", err)
		return
	}

	_, err = s.UserUpdate("", avatar)
	if err != nil {
		slog.Errorf("Error setting the avatar: %v", err)
		return
	}

	d.lastChangeAvatarTime = time.Now()
}
