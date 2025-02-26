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

// loggingWriter wraps gin.ResponseWriter to capture the total number of bytes written.
type loggingWriter struct {
	gin.ResponseWriter
	totalSize int
}

func (lw *loggingWriter) Write(data []byte) (int, error) {
	n, err := lw.ResponseWriter.Write(data)
	lw.totalSize += n
	return n, err
}

func (lw *loggingWriter) Size() int {
	// Return the recorded size if greater than or equal to zero.
	return lw.totalSize
}

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
			val = fmt.Sprintf("%g", float64(latency.Nanoseconds()))
		case "mcs":
			val = fmt.Sprintf("%g", float64(latency.Nanoseconds())/1000)
		case "ms":
			val = fmt.Sprintf("%g", float64(latency.Nanoseconds())/1000/1000)
		case "s":
			val = fmt.Sprintf("%g", float64(latency.Nanoseconds())/1000/1000/1000)
		default:
			ns := float64(latency.Nanoseconds())
			if ns >= 1000*1000*1000 {
				val = fmt.Sprintf("%gs", ns/1000/1000/1000)
			} else if ns >= 1000*1000 {
				val = fmt.Sprintf("%gms", ns/1000/1000)
			} else if ns >= 1000 {
				val = fmt.Sprintf("%gμs", ns/1000)
			} else {
				val = fmt.Sprintf("%gns", ns)
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
			val = fmt.Sprintf("%g", float64(size)/1024)
		case "mb":
			val = fmt.Sprintf("%g", float64(size)/(1024*1024))
		case "gb":
			val = fmt.Sprintf("%g", float64(size)/(1024*1024*1024))
		default:
			if size >= 1024*1024*1024 {
				val = fmt.Sprintf("%gGB", float64(size)/(1024*1024*1024))
			} else if size >= 1024*1024 {
				val = fmt.Sprintf("%gMB", float64(size)/(1024*1024))
			} else if size >= 1024 {
				val = fmt.Sprintf("%gKB", float64(size)/1024)
			} else {
				val = fmt.Sprintf("%dB", size)
			}
		}
	case "response_size":
		size := c.Writer.Size()
		switch strings.ToLower(unitSpec) {
		case "kb":
			val = fmt.Sprintf("%g", float64(size)/1024)
		case "mb":
			val = fmt.Sprintf("%g", float64(size)/(1024*1024))
		case "gb":
			val = fmt.Sprintf("%g", float64(size)/(1024*1024*1024))
		default:
			if size >= 1024*1024*1024 {
				val = fmt.Sprintf("%gGB", float64(size)/(1024*1024*1024))
			} else if size >= 1024*1024 {
				val = fmt.Sprintf("%gMB", float64(size)/(1024*1024))
			} else if size >= 1024 {
				val = fmt.Sprintf("%gKB", float64(size)/1024)
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

// FormatLogMessage constructs the log message using the globalLogFormat.
func FormatLogMessage(c *gin.Context, latency time.Duration) string {
	format := globalLogFormat
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

// LoggerMiddleware wraps the ResponseWriter and logs after the response is finished.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Wrap ResponseWriter to capture size.
		lw := &loggingWriter{ResponseWriter: c.Writer}
		c.Writer = lw
		start := time.Now()
		c.Next()
		// Force flush headers.
		c.Writer.WriteHeaderNow()
		latency := time.Since(start)
		msg := FormatLogMessage(c, latency)
		fmt.Println(msg)
		if len(c.Errors) > 0 {
			fmt.Println("api error:", c.Errors.String())
		}
	}
}
