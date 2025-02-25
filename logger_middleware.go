package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// A list of possible placeholders for custom log formatting.
var placeholders = []string{
	"{{time}}",
	"{{status_code}}",
	"{{method}}",
	"{{path}}",
	"{{client_ip}}",
	"{{latency}}",
	"{{headers}}",
	"{{cookies}}",
}

// generateRandomLogFormat automatically creates a random log format string by
// randomly shuffling placeholders and inserting random separators.
func generateRandomLogFormat() string {
	// Shuffle a copy of the placeholders slice.
	phCopy := make([]string, len(placeholders))
	copy(phCopy, placeholders)
	rand.Shuffle(len(phCopy), func(i, j int) {
		phCopy[i], phCopy[j] = phCopy[j], phCopy[i]
	})
	// Decide on a random number of segments (between 3 and 6)
	n := rand.Intn(4) + 3
	segments := []string{}
	for i := 0; i < n && i < len(phCopy); i++ {
		segments = append(segments, phCopy[i])
		if i < n-1 {
			// Append a random separator (with surrounding spaces)
			segments = append(segments, " "+randomSeparator()+" ")
		}
	}
	return strings.Join(segments, "")
}

func randomSeparator() string {
	seps := []string{"-", "|", ":", "~", "/", " "}
	return seps[rand.Intn(len(seps))]
}

// getLogFormat returns the log format string based on the LOG_FORMAT environment variable.
// If LOG_FORMAT is "RANDOM", it automatically generates a random format.
func getLogFormat() string {
	format := viper.GetString("LOG_FORMAT")
	format = strings.TrimSpace(format)
	if format == "" {
		// Default to JSON structured logging.
		return "json"
	}
	if strings.ToUpper(format) == "RANDOM" {
		return generateRandomLogFormat()
	}
	return format
}

// formatLogMessage formats the log message based on the given format string and request context.
func formatLogMessage(c *gin.Context, latency time.Duration, status int) string {
	// Prepare placeholder replacements.
	replacements := map[string]string{
		"{{time}}":        time.Now().UTC().Format(time.RFC3339Nano),
		"{{status_code}}": fmt.Sprintf("%d", status),
		"{{method}}":      c.Request.Method,
		"{{path}}":        c.Request.URL.Path,
		"{{client_ip}}":   c.ClientIP(),
		"{{latency}}":     latency.String(),
		"{{headers}}":     fmt.Sprintf("%v", c.Request.Header),
	}

	var cookies []string
	for _, cookie := range c.Request.Cookies() {
		cookies = append(cookies, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	replacements["{{cookies}}"] = strings.Join(cookies, "; ")

	// Retrieve the log format.
	format := getLogFormat()
	if format == "json" {
		// In JSON mode, we don't need to produce a custom string.
		return ""
	}

	// Replace placeholders in the custom format.
	logMsg := format
	for placeholder, value := range replacements {
		logMsg = strings.ReplaceAll(logMsg, placeholder, value)
	}
	return logMsg
}

// ZapLoggerMiddleware is a Gin middleware that logs API requests using the Zap logger.
// It uses the LOG_FORMAT environment variable to determine the log format.
func ZapLoggerMiddleware() gin.HandlerFunc {
	// Disable Gin debug messages by setting Gin to release mode.
	gin.SetMode(gin.ReleaseMode)

	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next() // Process the request.
		latency := time.Since(startTime)
		status := c.Writer.Status()

		// Retrieve the desired log format.
		format := getLogFormat()
		if format == "json" {
			// Structured JSON logging.
			logger.Info("api request",
				zap.Int("status", status),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.Duration("latency", latency),
			)
		} else {
			// Custom formatted log message.
			logMsg := formatLogMessage(c, latency, status)
			logger.Info(logMsg)
		}

		if len(c.Errors) > 0 {
			logger.Error("api error", zap.Any("errors", c.Errors))
		}
	}
}
