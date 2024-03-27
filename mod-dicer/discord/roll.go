package discord

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"

	"github.com/keshon/dice-roller/mod-dicer/utils"
)

// handleRollCommand handles the roll command for Discord.
func (d *Discord) handleRollCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {
	d.changeAvatar(s)

	if param == "" {
		param = "1d20"
	}

	tokens := tokenize(param)

	if len(tokens) > 10 {
		s.ChannelMessageSend(m.ChannelID, "Error: you can roll up to 10 dice in a single command.")
		return
	}

	totalResult, detailedResults, hasSingleMultiplier, processedKeys, err := processDiceTokens(tokens)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	embedMsg := embed.NewEmbed().
		SetTitle(fmt.Sprintf("= %d", totalResult)).
		SetColor(0x9f00d4)

	for _, key := range processedKeys {
		results := detailedResults[key]

		if len(processedKeys) == 1 && hasSingleMultiplier {
			embedMsg.AddField("", "`"+key+"`").MakeFieldInline()
		} else {
			embedMsg.AddField(fmt.Sprintf("(%s)\n", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(results)), " + "), "[]")), "`"+key+"`").MakeFieldInline()
		}
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embedMsg.MessageEmbed)
}

// processDiceTokens processes the tokens and returns the total result, detailed results, and processed keys.
func processDiceTokens(tokens []string) (int, map[string][]int, bool, []string, error) {
	totalResult := 0
	detailedResults := make(map[string][]int)
	processedKeys := []string{}
	hasSingleMultiplier := true

	for i, token := range tokens {
		multiplier, diceSides, err := parseToken(token)
		if err != nil {
			return 0, nil, false, nil, fmt.Errorf("invalid input. %v", err)
		}

		slog.Infof("[%v] work with token is %v:", i, token)
		slog.Infof("[%v] ..multiplier for token %v is %v: ", i, token, multiplier)
		slog.Infof("[%v] ..diceSides for token %v is %v: ", i, token, diceSides)

		if diceSides <= 0 {
			return 0, nil, false, nil, fmt.Errorf("Invalid input. Dice sides must be greater than zero.")
		}

		results := make([]int, utils.AbsInt(multiplier))
		for j := 0; j < multiplier; j++ {
			rollResult, err := secureRandomInt(diceSides)

			if err != nil {
				slog.Errorf("Error generating secure random number: %v", err)
				return 0, nil, false, nil, err
			}

			results[j] = rollResult
			totalResult += rollResult
		}

		key := fmt.Sprintf("%dd%d", multiplier, diceSides)
		detailedResults[key] = results
		processedKeys = append(processedKeys, key)

		if multiplier > 1 {
			hasSingleMultiplier = false
		}

		slog.Infof("[%v] ..rolled values are %v\n", i, results)
	}

	return totalResult, detailedResults, hasSingleMultiplier, processedKeys, nil
}

// secureRandomInt generates a secure random integer between 1 and max (inclusive).
func secureRandomInt(max int) (int, error) {
	if max <= 0 {
		return 0, nil
	}

	// Generate a random number in the range [0, max-1]
	randomNum, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}

	// Add 1 to make the range [1, max]
	return int(randomNum.Int64()) + 1, nil
}

// getRandomDeviation generates a random deviation based on initial velocity, angular velocity, air resistance, mass, and shape factor.
func getRandomDeviation(initialVelocity, angularVelocity, airResistance, mass, shapeFactor float64) float64 {
	// Simulate the effect of initial velocity, angular velocity, air resistance, mass, and shape factor on deviation
	combinedParameter := (math.Pow(initialVelocity, 0.8) * math.Pow(angularVelocity, 0.5) / airResistance) / (mass * shapeFactor)
	return combinedParameter - 0.5
}

// tokenize breaks down the input into tokens along with signs.
func tokenize(input string) []string {
	delimiters := `[^0-9d]+`
	r := regexp.MustCompile(delimiters)
	tokens := r.Split(input, -1)
	return tokens
}

// parseToken extracts multiplier and dice sides from a token.
func parseToken(token string) (int, int, error) {
	parts := strings.Split(token, "d")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("Invalid token format")
	}

	// Parse multiplier
	multiplier := 1
	if parts[0] != "" {
		var err error
		multiplier, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("Invalid multiplier in token: %s", token)
		}

		// Check for a reasonable range for multiplier (adjust the limit as needed)
		if multiplier <= 0 || multiplier > 10 {
			return 0, 0, fmt.Errorf("Multiplier should be between 1 and 10")
		}
	}

	// Parse dice sides
	diceSides, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("Invalid dice sides in token: %s", token)
	}

	// Check for a reasonable range for dice sides (adjust the limit as needed)
	if diceSides <= 0 || diceSides > 100 {
		return 0, 0, fmt.Errorf("Dice sides should be between 1 and 100")
	}

	return multiplier, diceSides, nil
}
