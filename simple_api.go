package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// defaultColor is selected at application startup.
var defaultColor string

func init() {
	// Initialize defaultColor from env variable RANDOM_HTML_API_COLOR if provided,
	// otherwise, generate a random color.
	envColor := viper.GetString("RANDOM_HTML_API_COLOR")
	if envColor != "" {
		defaultColor = envColor
	} else {
		defaultColor = randomColor()
	}
}

// randomColor returns a random hex color string.
func randomColor() string {
	letters := []rune("0123456789ABCDEF")
	color := "#"
	for i := 0; i < 6; i++ {
		color += string(letters[rand.Intn(len(letters))])
	}
	return color
}

// SimpleHandler handles GET /simple.
// Responds with "ok".
func SimpleHandler(c *gin.Context) {
	ResponseJSON(c, http.StatusOK, gin.H{"message": "ok"})
}

// FooHandler handles GET /simple/foo.
// Responds with "foo ok" along with request header details.
func FooHandler(c *gin.Context) {
	details := getRequestDetails(c)
	details["message"] = "foo ok"
	ResponseJSON(c, http.StatusOK, details)
}

// BarHandler handles POST /simple/bar.
// Responds with "bar ok" and includes parsed request headers and body info.
func BarHandler(c *gin.Context) {
	var body interface{}
	// Attempt to bind the JSON body. On error, body remains nil.
	c.ShouldBindJSON(&body)
	details := getRequestDetails(c)
	details["body"] = gin.H{
		"payload": body,
	}
	details["message"] = "bar ok"
	ResponseJSON(c, http.StatusOK, details)
}

// ColorHandler handles GET /simple/color?color=[string] and returns HTML (not JSON).
// It uses the provided query parameter "color" (processed with RANDOM syntax if needed)
// or falls back to the RANDOM_HTML_API_COLOR env variable / defaultColor.
func ColorHandler(c *gin.Context) {
	// Check query parameter "color"
	color := c.Query("color")
	if color != "" {
		// Process RANDOM syntax if provided.
		processed, err := processRandomValue(color)
		if err == nil {
			if s, ok := processed.(string); ok {
				color = s
			}
		}
	} else {
		// If not provided, try env variable or default.
		envColor := viper.GetString("RANDOM_HTML_API_COLOR")
		if envColor != "" {
			color = envColor
		} else {
			color = defaultColor
		}
	}

	// Build HTML response that includes request details.
	details := getRequestDetails(c)
	var detailsStr strings.Builder
	for key, value := range details {
		detailsStr.WriteString(fmt.Sprintf("<p>%s: %v</p>", key, value))
	}
	// Include the current timestamp.
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	html := fmt.Sprintf(`
		<html>
		<head><title>Random Color API</title></head>
		<body style="background-color:%s;">
			<h1>Color API</h1>
			%s
			<p>requested_at: %s</p>
		</body>
		</html>
	`, color, detailsStr.String(), timestamp)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// LargeHandler handles GET /simple/large?length=<number>&sentence=[string].
// It repeats the provided sentence (or a default sentence) length times.
func LargeHandler(c *gin.Context) {
	// Parse "length" query parameter.
	lengthStr := c.Query("length")
	length, err := strconv.Atoi(lengthStr)
	if err != nil || length <= 0 {
		length = 10 // default repetition count.
	}
	sentence := c.Query("sentence")
	if sentence == "" {
		sentence = "This is a sample sentence."
	}
	// Process RANDOM syntax for sentence if provided.
	processed, err := processRandomValue(sentence)
	if err == nil {
		if s, ok := processed.(string); ok {
			sentence = s
		}
	}
	// Build large text by repeating the sentence.
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteString(sentence)
		if i < length-1 {
			sb.WriteString(" ")
		}
	}
	ResponseJSON(c, http.StatusOK, gin.H{"large_text": sb.String()})
}
