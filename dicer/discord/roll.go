package discord

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"

	"github.com/keshon/dice-roller/dicer/utils"
)

// handleRollCommand handles the roll command for Discord.
func (d *Discord) handleRollCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {
	d.changeAvatar(s)

	slog.Warn("user input is", param)

	// Tokenization logic
	tokens := tokenize(param)

	slog.Warn("all tokens are", tokens)

	// Initialize variables for the rolling logic
	totalResult := 0
	detailedResults := make(map[string][]int)
	processedKeys := []string{} // Keep track of processed keys
	isOneMultiplier := true

	// Rolling logic
	for i, token := range tokens {
		// Parse each token
		multiplier, diceSides, err := parseToken(token)

		slog.Infof("[%v] work with token is %v:", i, token)
		slog.Infof("[%v] ..multiplier for token %v is %v: ", i, token, multiplier)
		slog.Infof("[%v] ..diceSides for token %v is %v: ", i, token, diceSides)

		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid input. Please use a valid dice expression, e.g., `1d20`.")
			return
		}

		// Check for valid dice sides
		if diceSides <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Invalid input. Dice sides must be greater than zero.")
			return
		}

		// Roll the dice
		results := make([]int, utils.AbsInt(multiplier))
		for j := 0; j < multiplier; j++ {
			rollResult := rand.Intn(diceSides) + 1
			slog.Infof("[%v] ..rolled value is %v\n", i, rollResult)
			results[j] = rollResult
			totalResult += rollResult
		}

		// Store detailed results
		key := fmt.Sprintf("%dd%d", multiplier, diceSides)
		detailedResults[key] = results
		processedKeys = append(processedKeys, key)

		if multiplier > 1 {
			isOneMultiplier = false
		}
	}

	// Sending result to the Discord channel
	embedMsg := embed.NewEmbed().
		SetTitle(fmt.Sprintf("= %d", totalResult)).
		// SetDescription(fmt.Sprintf("Roll for`%s`:\n", param)).
		// SetFooter(fmt.Sprintf("Roll for %s:\n", param)).
		SetColor(0x9f00d4)

	// Formatting detailed results in the correct order
	for _, key := range processedKeys {
		results := detailedResults[key]

		// Check if only one dice was requested and multiplier is also 1
		if len(tokens) == 1 && isOneMultiplier {
			// If only one dice with multiplier 1, don't include the value in the field
			embedMsg.AddField("", "`"+key+"`").MakeFieldInline()
		} else {
			embedMsg.AddField(fmt.Sprintf("(%s)\n", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(results)), " + "), "[]")), "`"+key+"`").MakeFieldInline()
		}
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embedMsg.MessageEmbed)
}

// tokenize breaks down the input into tokens along with signs.
func tokenize(input string) []string {
	// Define regular expression to split input based on delimiters
	delimiters := `[^0-9d]+`
	r := regexp.MustCompile(delimiters)

	// Use regular expression to split input based on delimiters
	tokens := r.Split(input, -1)

	return tokens
}

// parseToken extracts multiplier and dice sides from a token.
func parseToken(token string) (int, int, error) {
	parts := strings.Split(token, "d")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("Invalid token format")
	}

	multiplier := 1
	if parts[0] != "" {
		multiplier, _ = strconv.Atoi(parts[0])
	}

	diceSides, _ := strconv.Atoi(parts[1])

	return multiplier, diceSides, nil
}
