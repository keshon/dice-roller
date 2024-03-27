package utils

import (
	"encoding/base64"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// FormatDuration formats the given seconds into HH:MM:SS format.
// Example: formattedTime := FormatDuration(3661.5) // Returns "01:01:02"
func FormatDuration(seconds float64) string {
	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	totalSeconds %= 3600
	minutes := totalSeconds / 60
	seconds = math.Mod(float64(totalSeconds), 60)
	return fmt.Sprintf("%02d:%02d:%02.0f", hours, minutes, seconds)
}

// ReadFileToBase64 reads a file and returns its base64 representation with data URI.
// Example: base64Data, err := ReadFileToBase64("/path/to/image.jpg")
func ReadFileToBase64(filePath string) (string, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading the file: %v", err)
	}

	base64Content := base64.StdEncoding.EncodeToString(fileContent)
	return fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(fileContent), base64Content), nil
}

// SanitizeString removes unwanted characters from the input string.
// Example: sanitizedStr := SanitizeString("Hello#World!")
func SanitizeString(input string) string {
	unwantedCharRegex := regexp.MustCompile("[[:^print:]]")
	return unwantedCharRegex.ReplaceAllString(input, "")
}

// InferProtocolByPort attempts to infer the protocol based on the availability of a specific port.
// Example: protocol := InferProtocolByPort("example.com", 443)
func InferProtocolByPort(hostname string, port int) string {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		// Assuming it's not available, default to HTTP
		return "http://"
	}
	defer conn.Close()

	// The port is available, use HTTPS
	return "https://"
}

func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func parseFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

// GetWeightedRandomImagePath returns a random image path with reduced chances for recently shown images.
func GetWeightedRandomImagePath(folderPath string) (string, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return "", err
	}

	var images []string
	var weights []int

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			images = append(images, file.Name())
			weights = append(weights, 1) // Initial weight for each image is 1
		}
	}

	if len(images) == 0 {
		return "", fmt.Errorf("no valid images found")
	}

	totalWeights := len(images) // Initial total weights equal to the number of images

	randomWeight := rand.Intn(totalWeights)

	index := -1
	for i, weight := range weights {
		if randomWeight < weight {
			index = i
			break
		}
		randomWeight -= weight
	}

	if index == -1 {
		return "", fmt.Errorf("error selecting random image")
	}

	// Decrease the weight of the recently selected image
	weights[index] = weights[index] / 2

	// Increase the weight of all other images
	for i := range weights {
		if i != index {
			weights[i] = weights[i] * 2
		}
	}

	imagePath := filepath.Join(folderPath, images[index])
	return imagePath, nil
}

// TrimString trims the string's ending beyond the specified character limit.
// Example: trimmedText := TrimString("This is a long text.", 10) // Returns "This is a"
func TrimString(input string, limit int) string {
	if len(input) <= limit {
		return input
	}

	return input[:limit]
}

func AbsInt(x int) int {
	return absDiffInt(x, 0)
}

func absDiffInt(x, y int) int {
	if x < y {
		return y - x
	}
	return x - y
}

func absDiffUint(x, y uint) uint {
	if x < y {
		return y - x
	}
	return x - y
}
