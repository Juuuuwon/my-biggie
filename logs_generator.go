package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// LogsGeneratorPayload defines the JSON payload for generating logs.
type LogsGeneratorPayload struct {
	MaintainSecond      DuckInt `json:"maintain_second"`
	LogCountPerInterval DuckInt `json:"log_count_per_interval"`
	LinePerLog          DuckInt `json:"line_per_log"`
	IntervalSeconds     DuckInt `json:"interval_seconds"`
	Async               bool    `json:"async"`
}

// generateRandomLog creates a log message using random values and the LOG_FORMAT configuration.
// Supported placeholders: {{time}}, {{status_code}}, {{method}}, {{path}}, {{client_ip}}, {{latency}}, {{cookies}}
func generateRandomLog() string {
	now := time.Now().UTC().Format(time.RFC3339)
	statusCodes := []int{200, 201, 400, 401, 404, 500}
	methodList := []string{"GET", "POST", "PUT", "DELETE"}
	paths := []string{"/api/random", "/test", "/stress", "/metrics"}
	clientIPs := []string{"192.168.1.1", "10.0.0.5", "172.16.0.3"}
	latency := time.Duration(rand.Intn(500)) * time.Millisecond
	cookies := "[cookie1=value1; cookie2=value2]"

	randomStatus := fmt.Sprintf("%d", statusCodes[rand.Intn(len(statusCodes))])
	randomMethod := methodList[rand.Intn(len(methodList))]
	randomPath := paths[rand.Intn(len(paths))]
	randomIP := clientIPs[rand.Intn(len(clientIPs))]

	format := viper.GetString("LOG_FORMAT")
	if format == "" {
		format = "json"
	}
	// If LOG_FORMAT is RANDOM, generate a random format string.
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
		// Shuffle the slice.
		rand.Shuffle(len(placeholders), func(i, j int) {
			placeholders[i], placeholders[j] = placeholders[j], placeholders[i]
		})
		format = strings.Join(placeholders, " | ")
	} else if strings.EqualFold(format, "apache") {
		format = "{{client_ip}} - - [{{time}}] \"{{method}} {{path}} HTTP/1.1\" {{status_code}} -"
	} else if strings.EqualFold(format, "nginx") {
		format = "{{client_ip}} - [{{time}}] \"{{method}} {{path}}\" {{status_code}} {{latency}}"
	} else if strings.EqualFold(format, "json") {
		// Return JSON formatted log
		logObj := map[string]interface{}{
			"time":        now,
			"status_code": randomStatus,
			"method":      randomMethod,
			"path":        randomPath,
			"client_ip":   randomIP,
			"latency":     latency.String(),
			"cookies":     cookies,
		}
		if b, err := json.Marshal(logObj); err == nil {
			return string(b)
		}
	}
	// For custom formats, replace placeholders.
	replacements := map[string]string{
		"{{time}}":        now,
		"{{status_code}}": randomStatus,
		"{{method}}":      randomMethod,
		"{{path}}":        randomPath,
		"{{client_ip}}":   randomIP,
		"{{latency}}":     latency.String(),
		"{{cookies}}":     cookies,
	}
	msg := format
	for placeholder, value := range replacements {
		msg = strings.ReplaceAll(msg, placeholder, value)
	}
	return msg
}

// LogsGeneratorHandler handles POST /stress/logs.
// It generates logs over time using random log content based on the current LOG_FORMAT.
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
				// Generate each log with multiple lines using random log content.
				for j := 0; j < linePerLog; j++ {
					lines = append(lines, generateRandomLog())
				}
				logMessage := strings.Join(lines, "\n")
				// Log the message using our fmt-based logger.
				log(logMessage)
			}
			time.Sleep(interval)
		}
		log("Logs generation completed", "duration_sec", maintainSec)
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":                "Logs generation started",
			"maintain_second":        maintainSec,
			"log_count_per_interval": logCountPerInterval,
			"line_per_log":           linePerLog,
			"interval_seconds":       intervalSec,
		})
	} else {
		stressFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":                "Logs generation completed",
			"maintain_second":        maintainSec,
			"log_count_per_interval": logCountPerInterval,
			"line_per_log":           linePerLog,
			"interval_seconds":       intervalSec,
		})
	}
}
