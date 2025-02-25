package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// FormatLogMessage builds a log message string based on the LOG_FORMAT environment variable.
// Supported placeholders:
//
//	{{time}}, {{status_code}}, {{method}}, {{path}}, {{client_ip}}, {{latency}}, {{cookies}}.
func FormatLogMessage(c *gin.Context, latency time.Duration) string {
	format := viper.GetString("LOG_FORMAT")
	if format == "" {
		format = "json"
	}
	// If format is RANDOM, generate a random format string.
	if strings.EqualFold(format, "RANDOM") {
		placeholders := []string{
			"[{{time}}]",
			"{{status_code}}",
			"{{method}}",
			"{{path}}",
			"{{client_ip}}",
			"latency: {{latency}}",
			"cookies: {{cookies}}",
		}
		randOrder := make([]string, len(placeholders))
		copy(randOrder, placeholders)
		// Shuffle the slice.
		for i := range randOrder {
			j := i + int(time.Now().UnixNano()%int64(len(randOrder)-i))
			randOrder[i], randOrder[j] = randOrder[j], randOrder[i]
		}
		format = strings.Join(randOrder, " | ")
	}
	// Predefined formats.
	if strings.EqualFold(format, "apache") {
		format = "{{client_ip}} - - [{{time}}] \"{{method}} {{path}} HTTP/1.1\" {{status_code}} -"
	} else if strings.EqualFold(format, "nginx") {
		format = "{{client_ip}} - [{{time}}] \"{{method}} {{path}}\" {{status_code}} {{latency}}"
	} else if strings.EqualFold(format, "json") {
		logObj := map[string]interface{}{
			"time":        time.Now().UTC().Format(time.RFC3339),
			"status_code": c.Writer.Status(),
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"client_ip":   c.ClientIP(),
			"latency":     latency.String(),
			"cookies":     c.Request.Cookies(),
		}
		b, err := json.Marshal(logObj)
		if err == nil {
			return string(b)
		}
	}
	// For custom formats, replace placeholders.
	replacements := map[string]string{
		"{{time}}":        time.Now().UTC().Format(time.RFC3339),
		"{{status_code}}": fmt.Sprintf("%d", c.Writer.Status()),
		"{{method}}":      c.Request.Method,
		"{{path}}":        c.Request.URL.Path,
		"{{client_ip}}":   c.ClientIP(),
		"{{latency}}":     latency.String(),
		"{{cookies}}":     fmt.Sprintf("%v", c.Request.Cookies()),
	}
	msg := format
	for placeholder, value := range replacements {
		msg = strings.ReplaceAll(msg, placeholder, value)
	}
	return msg
}

// LoggerMiddleware logs each API request using our fmt-based logging functions.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		latency := time.Since(startTime)
		msg := FormatLogMessage(c, latency)
		log(msg)
		if len(c.Errors) > 0 {
			log("api error", "errors", c.Errors.String())
		}
	}
}
