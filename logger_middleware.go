package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// placeholderRegex matches substrings like {<placeholder>} or {<placeholder>:<unit>}
var placeholderRegex = regexp.MustCompile(`\{([^}]+)\}`)

// resolvePlaceholder processes a single placeholder (e.g., "latency:ms" or "time:%Y-%m-%dT%H:%M:%S")
// and returns its string representation using actual request values.
func resolvePlaceholder(content string, c *gin.Context, latency time.Duration) (string, error) {
	parts := strings.SplitN(content, ":", 2)
	key := strings.ToLower(strings.TrimSpace(parts[0]))
	unitSpec := ""
	if len(parts) == 2 {
		unitSpec = strings.TrimSpace(parts[1])
	}
	var val string
	switch key {
	case "time":
		now := time.Now().UTC()
		if unitSpec != "" {
			layout := convertTimeFormat(unitSpec)
			val = now.Format(layout)
		} else {
			val = now.Format(time.RFC3339)
		}
	case "status_code":
		val = strconv.Itoa(c.Writer.Status())
	case "method":
		val = c.Request.Method
	case "path":
		val = c.Request.URL.Path
	case "client_ip":
		val = c.ClientIP()
	case "latency":
		switch strings.ToLower(unitSpec) {
		case "ns":
			val = strconv.FormatInt(latency.Nanoseconds(), 10)
		case "ms":
			val = strconv.FormatInt(latency.Milliseconds(), 10)
		case "s":
			val = fmt.Sprintf("%.2f", latency.Seconds())
		default:
			// If no unit provided, convert to a human-readable unit.
			ms := float64(latency.Milliseconds())
			if ms >= 1000 {
				val = fmt.Sprintf("%.3fs", ms/1000)
			} else {
				val = fmt.Sprintf("%.3fms", ms)
			}
		}
	case "user_agent":
		val = c.Request.UserAgent()
	case "protocol":
		val = c.Request.Proto
	case "request_size":
		size := c.Request.ContentLength
		switch strings.ToLower(unitSpec) {
		case "kb":
			val = fmt.Sprintf("%.3f", float64(size)/1024)
		case "mb":
			val = fmt.Sprintf("%.3f", float64(size)/(1024*1024))
		case "gb":
			val = fmt.Sprintf("%.3f", float64(size)/(1024*1024*1024))
		default:
			// Human-readable conversion:
			if size >= 1024*1024*1024 {
				val = fmt.Sprintf("%.3fGB", float64(size)/(1024*1024*1024))
			} else if size >= 1024*1024 {
				val = fmt.Sprintf("%.3fMB", float64(size)/(1024*1024))
			} else if size >= 1024 {
				val = fmt.Sprintf("%.3fKB", float64(size)/1024)
			} else {
				val = fmt.Sprintf("%dB", size)
			}
		}
	case "response_size":
		size := c.Writer.Size()
		switch strings.ToLower(unitSpec) {
		case "kb":
			val = fmt.Sprintf("%.3f", float64(size)/1024)
		case "mb":
			val = fmt.Sprintf("%.3f", float64(size)/(1024*1024))
		case "gb":
			val = fmt.Sprintf("%.3f", float64(size)/(1024*1024*1024))
		default:
			if size >= 1024*1024*1024 {
				val = fmt.Sprintf("%.3fGB", float64(size)/(1024*1024*1024))
			} else if size >= 1024*1024 {
				val = fmt.Sprintf("%.3fMB", float64(size)/(1024*1024))
			} else if size >= 1024 {
				val = fmt.Sprintf("%.3fKB", float64(size)/1024)
			} else {
				val = fmt.Sprintf("%dB", size)
			}
		}
	default:
		return "", fmt.Errorf("unsupported placeholder: %s", key)
	}
	return val, nil
}

// convertTimeFormat converts a strftime-like format to Go time layout.
func convertTimeFormat(format string) string {
	replacements := map[string]string{
		"%Y": "2006",
		"%m": "01",
		"%d": "02",
		"%H": "15",
		"%M": "04",
		"%S": "05",
	}
	result := format
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	return result
}

// FormatLogMessage constructs the log message using globalLogFormat.
func FormatLogMessage(c *gin.Context, latency time.Duration) string {
	format := globalLogFormat
	// For predefined formats, we assume they are custom defined in globalLogFormat.
	// Replace all placeholders in the format.
	result := placeholderRegex.ReplaceAllStringFunc(format, func(match string) string {
		content := strings.Trim(match, "{}")
		val, err := resolvePlaceholder(content, c, latency)
		if err != nil {
			return "ERR"
		}
		return val
	})
	return result
}

// LoggerMiddleware logs each API request using the global log format.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		msg := FormatLogMessage(c, latency)
		fmt.Println(msg)
		if len(c.Errors) > 0 {
			fmt.Println("api error:", c.Errors.String())
		}
	}
}
