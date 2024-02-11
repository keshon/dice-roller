package discord

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
)

// handleRollCommand handles the roll command for Discord.
func (d *Discord) handleRollCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {
	d.changeAvatar(s)

	// Tokenization logic
	tokens, signs := tokenize(param)
	slog.Warn("Tokens are", tokens)
	slog.Warn("Signs are", signs)

	// Initialize variables for the rolling logic
	totalResult := 0
	detailedResults := make(map[string][]int)
	processedKeys := []string{} // Keep track of processed keys

	// Rolling logic
	for i, token := range tokens {
		// Parse each token
		multiplier, diceSides, err := parseToken(token)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid input. Please use a valid dice expression, e.g., `1d20`.")
			return
		}

		// Check for valid dice sides
		if diceSides <= 0 {
			s.ChannelMessageSend(m.ChannelID, "Invalid input. Dice sides must be greater than zero.")
			return
		}

		// Determine sign
		sign := "+"
		if i < len(signs) {
			sign = signs[i]
		}

		// Roll the dice
		results := make([]int, multiplier)
		for j := 0; j < multiplier; j++ {
			rollResult := rand.Intn(diceSides) + 1
			if sign == "-" {
				rollResult = -rollResult
			}
			results[j] = rollResult
			totalResult += rollResult
		}

		// Store detailed results
		var key string
		if sign == "+" {
			key = fmt.Sprintf("%dd%d", multiplier, diceSides)
		} else if sign == "-" {
			key = fmt.Sprintf("%dd%d", multiplier, diceSides)
			results = negateResults(results)
		}
		detailedResults[key] = results
		processedKeys = append(processedKeys, key)
	}

	// Formatting detailed results in the correct order
	detailedResultStr := ""
	for _, key := range processedKeys {
		results := detailedResults[key]
		detailedResultStr += fmt.Sprintf("%s: (%s)\n", key, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(results)), " + "), "[]"))
	}

	// Formatting final result
	resultStr := fmt.Sprintf("rolling %s\n%s= %d", param, detailedResultStr, totalResult)

	// Sending result to the Discord channel
	embedMsg := embed.NewEmbed().
		SetDescription(resultStr).
		SetColor(0x9f00d4).MessageEmbed

	s.ChannelMessageSendEmbed(m.ChannelID, embedMsg)
}

// negateResults negates each element in the results slice
func negateResults(results []int) []int {
	for i := range results {
		results[i] = -results[i]
	}
	return results
}

// tokenize breaks down the input into tokens along with signs.
func tokenize(input string) ([]string, []string) {
	// You can implement your own logic for tokenization based on the input format.
	// For simplicity, let's use a basic approach for now.
	input = strings.ReplaceAll(input, " ", "")

	var tokens []string
	var signs []string

	for _, char := range input {
		if char == '+' || char == '-' {
			signs = append(signs, string(char))
		}
	}

	tokens = strings.FieldsFunc(input, func(r rune) bool {
		return r == '+' || r == '-'
	})

	return tokens, signs
}

// parseToken extracts multiplier, dice sides, and constant from a token.
func parseToken(token string) (int, int, error) {
	// Your implementation for parsing each token.
	// For simplicity, let's assume the token format is "NdM+C" where N, M, and C are integers.
	parts := strings.Split(token, "d")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("Invalid token format")
	}

	multiplier, _ := strconv.Atoi(parts[0])

	// Check for constant term
	var diceSidesStr string
	if strings.Contains(parts[1], "+") {
		splitParts := strings.Split(parts[1], "+")
		diceSidesStr = splitParts[0]
	} else if strings.Contains(parts[1], "-") {
		splitParts := strings.Split(parts[1], "-")
		diceSidesStr = splitParts[0]
	} else {
		diceSidesStr = parts[1]
	}

	diceSides, _ := strconv.Atoi(diceSidesStr)

	slog.Info("Token", token)
	slog.Info("All Values detected", parts)
	slog.Info("Value detected", parts[1])
	slog.Info("Multiplier detected", multiplier)
	slog.Info("diceSides detected", diceSides)

	return multiplier, diceSides, nil
}
