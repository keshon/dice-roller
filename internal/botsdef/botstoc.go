package botsdef

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	about "github.com/keshon/dice-roller/mod-about/discord"
	dicer "github.com/keshon/dice-roller/mod-dicer/discord"
)

var Modules = []string{"about", "dicer"}

// CreateBotInstance creates a new bot instance based on the module name.
//
// Parameters:
// - session: a Discord session
// - module: the name of the module ("hi" or "hello")
// Returns a Discord instance.
func CreateBotInstance(session *discordgo.Session, module string) Discord {
	switch module {
	case "dicer":
		return dicer.NewDiscord(session)
	case "about":
		return about.NewDiscord(session)

	// ..add more cases for other modules if needed

	default:
		slog.Printf("Unknown module: %s", module)
		return nil
	}
}
