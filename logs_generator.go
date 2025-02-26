package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
)

// LogsGeneratorPayload defines the payload for generating fake log messages.
type LogsGeneratorPayload struct {
	MaintainSecond      DuckInt `json:"maintain_second"`
	LogCountPerInterval DuckInt `json:"log_count_per_interval"`
	LinePerLog          DuckInt `json:"line_per_log"`
	IntervalSeconds     DuckInt `json:"interval_seconds"`
	Async               bool    `json:"async"`
}

// GenerateRandomLogMessage creates a random log message using globalLogFormat
// and random values for each placeholder.
func GenerateRandomLogMessage() string {
	now := time.Now().UTC()
	// Generate random values for each placeholder.
	randomValues := map[string]string{
		"time":          now.Format(time.RFC3339),
		"status_code":   strconv.Itoa([]int{200, 201, 400, 401, 404, 500}[rand.Intn(6)]),
		"method":        []string{"GET", "POST", "PUT", "DELETE"}[rand.Intn(4)],
		"path":          []string{"/dummy", "/test", "/stress", "/metrics", "/api/data"}[rand.Intn(5)],
		"client_ip":     fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256)),
		"latency":       fmt.Sprintf("%dms", rand.Intn(500)+10),
		"user_agent":    []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", "curl/7.68.0", "PostmanRuntime/7.26.8"}[rand.Intn(4)],
		"protocol":      []string{"HTTP/1.1", "HTTP/2"}[rand.Intn(2)],
		"request_size":  strconv.Itoa(rand.Intn(9900) + 100),
		"response_size": strconv.Itoa(rand.Intn(9900) + 100),
		"cookies":       fmt.Sprintf("cookie1=value%d; cookie2=value%d", rand.Intn(1000), rand.Intn(1000)),
	}

	// Use globalLogFormat.
	format := globalLogFormat
	// Replace placeholders in the format.
	result := placeholderRegex.ReplaceAllStringFunc(format, func(match string) string {
		content := strings.Trim(match, "{}")
		parts := strings.SplitN(content, ":", 2)
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		unit := ""
		if len(parts) == 2 {
			unit = strings.TrimSpace(parts[1])
		}
		val, exists := randomValues[key]
		if !exists {
			return "ERR"
		}
		// For latency, request_size, response_size, apply unit conversion if unit specified.
		switch key {
		case "latency":
			// Assume original is in ms.
			msVal, _ := strconv.Atoi(strings.TrimSuffix(val, "ms"))
			switch strings.ToLower(unit) {
			case "ns":
				return strconv.FormatInt(int64(msVal)*1e6, 10)
			case "ms":
				return strconv.Itoa(msVal)
			case "s":
				return fmt.Sprintf("%.2f", float64(msVal)/1000)
			default:
				// Human readable with unit label.
				return fmt.Sprintf("%dms", msVal)
			}
		case "request_size", "response_size":
			numVal, _ := strconv.Atoi(val)
			switch strings.ToLower(unit) {
			case "kb":
				return fmt.Sprintf("%.3f", float64(numVal)/1024)
			case "mb":
				return fmt.Sprintf("%.3f", float64(numVal)/(1024*1024))
			case "gb":
				return fmt.Sprintf("%.3f", float64(numVal)/(1024*1024*1024))
			default:
				return fmt.Sprintf("%dB", numVal)
			}
		default:
			return val
		}
	})
	return result
}

// LogsGeneratorHandler handles POST /stress/logs.
// It generates random log messages using GenerateRandomLogMessage over time.
func LogsGeneratorHandler(c *gin.Context) {
	var payload LogsGeneratorPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	logCountPerInterval := int(payload.LogCountPerInterval)
	linePerLog := int(payload.LinePerLog)
	intervalSec := int(payload.IntervalSeconds)

	stressFunc := func() {
		endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
		interval := time.Duration(intervalSec) * time.Second
		for time.Now().Before(endTime) {
			for i := 0; i < logCountPerInterval; i++ {
				var lines []string
				for j := 0; j < linePerLog; j++ {
					lines = append(lines, GenerateRandomLogMessage())
				}
				combined := strings.Join(lines, "\n")
				// Print the log message.
				fmt.Println(combined)
			}
			time.Sleep(interval)
		}
		fmt.Println("Logs generation completed")
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, http.StatusOK, map[string]interface{}{
			"message":                "Logs generation started",
			"maintain_second":        maintainSec,
			"log_count_per_interval": logCountPerInterval,
			"line_per_log":           linePerLog,
			"interval_seconds":       intervalSec,
		})
	} else {
		stressFunc()
		ResponseJSON(c, http.StatusOK, map[string]interface{}{
			"message":                "Logs generation completed",
			"maintain_second":        maintainSec,
			"log_count_per_interval": logCountPerInterval,
			"line_per_log":           linePerLog,
			"interval_seconds":       intervalSec,
		})
	}
}
