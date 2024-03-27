package discord

import (
	"fmt"
	"os"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/dice-roller/internal/config"
	"github.com/keshon/dice-roller/internal/version"
	"github.com/keshon/dice-roller/mod-about/utils"
)

// handleHelpCommand handles the help command for the Discord bot.
//
// Takes in a session and a message create, and does not return any value.
func (d *Discord) handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	var hostname string
	if os.Getenv("HOST") == "" {
		hostname = config.RestHostname
	} else {
		hostname = os.Getenv("HOST") // from docker environment
	}

	avatarUrl := utils.InferProtocolByPort(hostname, 443) + hostname + "/avatar/random?" + fmt.Sprint(time.Now().UnixNano())
	slog.Info(avatarUrl)

	prefix := d.CommandPrefix

	rollShort := fmt.Sprintf("`%vroll` - default single roll of 1d20\n", prefix)
	rollFull := fmt.Sprintf("`%vroll 2d20` - single roll\n", prefix)
	rollMulti := fmt.Sprintf("`%vroll 1d20 2d6 1d4` - rolling several dice and adding up the result\n", prefix)
	help := fmt.Sprintf("**Show help**: `%vhelp` \nAliases: `%vh`\n", prefix, prefix)
	about := fmt.Sprintf("**Show version**: `%vabout`", prefix)
	register := fmt.Sprintf("**Enable commands listening**: `%vregister`\n", prefix)
	unregister := fmt.Sprintf("**Disable commands listening**: `%vunregister`", prefix)

	embedMsg := embed.NewEmbed().
		SetTitle("ℹ️ Dice Roller — Command Usage").
		SetDescription("Some commands are aliased for shortness.\n").
		AddField("", "*Rolls*\n"+rollShort+rollFull+rollMulti).
		AddField("", "").
		AddField("", "*General*\n"+help+about).
		AddField("", "").
		AddField("", "*Administration*\n"+register+unregister).
		SetThumbnail(avatarUrl). // TODO: move out to config .env file
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed

	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
}
