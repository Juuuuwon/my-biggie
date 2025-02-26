package main

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// globalLogFormat is the log format string used throughout the application.
var globalLogFormat string

// possiblePlaceholders (case-insensitive) that can be used in log format.
var requiredPlaceholders = []string{"time", "status_code", "method", "path", "client_ip"}
var optionalPlaceholders = []string{"latency", "user_agent", "protocol", "request_size", "response_size"}

// generateRandomTimeFormat generates a random strftime format for time.
// It must include %Y, %m, %d, %H, %M, %S.
func generateRandomTimeFormat() string {
	// For simplicity, we randomly choose a separator from a set.
	separators := []string{"-", "/", ":", "."}
	sep := separators[rand.Intn(len(separators))]
	// Force a format with year, month, day, hour, minute, second.
	return "%Y" + sep + "%m" + sep + "%d" + "T" + "%H" + sep + "%M" + sep + "%S"
}

// generateRandomGlobalLogFormat creates a random log format meeting the following rules:
// - Must include required placeholders: time, status_code, method, path, client_ip.
// - Not more than 7 placeholders total.
// - Randomly choose some optional placeholders (0 up to 2 additional ones).
// - Placeholders are separated by a single space.
// - Each placeholder can optionally have a unit specifier (for time, latency, request_size, response_size).
// - Randomly quote some placeholders with "", ” or [].
// - Randomly insert " - " between placeholders.
func generateRandomGlobalLogFormat() string {
	// Start with required placeholders.
	placeholders := make([]string, len(requiredPlaceholders))
	copy(placeholders, requiredPlaceholders)

	// Randomly decide how many optional placeholders to add (0 to 2)
	optCount := rand.Intn(3)
	for i := 0; i < optCount; i++ {
		opt := optionalPlaceholders[rand.Intn(len(optionalPlaceholders))]
		// Ensure we don't duplicate.
		already := false
		for _, p := range placeholders {
			if strings.EqualFold(p, opt) {
				already = true
				break
			}
		}
		if !already && len(placeholders) < 7 {
			placeholders = append(placeholders, opt)
		}
	}
	// Shuffle placeholders randomly.
	rand.Shuffle(len(placeholders), func(i, j int) {
		placeholders[i], placeholders[j] = placeholders[j], placeholders[i]
	})
	// For each placeholder, randomly decide to add a unit specifier.
	for i, ph := range placeholders {
		// For time, latency, request_size, response_size.
		switch strings.ToLower(ph) {
		case "time":
			// 50% chance to add a random time format.
			if rand.Float64() < 0.5 {
				placeholders[i] = fmt.Sprintf("{time:%s}", generateRandomTimeFormat())
			} else {
				placeholders[i] = "{time}"
			}
		case "latency":
			// Randomly choose from s, ms, micros, ns.
			units := []string{"s", "ms", "micros", "ns"}
			placeholders[i] = fmt.Sprintf("{latency:%s}", units[rand.Intn(len(units))])
		case "request_size", "response_size":
			// Randomly choose from b, kb, mb, gb.
			units := []string{"b", "kb", "mb", "gb"}
			placeholders[i] = fmt.Sprintf("{%s:%s}", strings.ToLower(ph), units[rand.Intn(len(units))])
		default:
			// Leave others as is.
			placeholders[i] = fmt.Sprintf("{%s}", ph)
		}
		// Randomly, with 50% chance, wrap the placeholder with quotes, single quotes, or square brackets.
		if rand.Float64() < 0.5 {
			wrappers := []struct{ open, close string }{
				{`"`, `"`},
				{"'", "'"},
				{"[", "]"},
			}
			w := wrappers[rand.Intn(len(wrappers))]
			placeholders[i] = w.open + placeholders[i] + w.close
		}
	}
	// Build the final format string by joining placeholders with spaces, and randomly insert " - " between some.
	var parts []string
	for _, ph := range placeholders {
		parts = append(parts, ph)
		// With 30% chance, append a " - " as a separate token.
		if rand.Float64() < 0.3 {
			parts = append(parts, "-")
		}
	}
	return strings.Join(parts, " ")
}

// initConfig reads configuration and sets defaults, including globalLogFormat.
func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	_ = viper.ReadInConfig() // ignore error, use defaults if no file
	viper.AutomaticEnv()     // read environment variables
	viper.SetDefault("LOG_FORMAT", "apache")

	logFormat := viper.GetString("LOG_FORMAT")
	switch strings.ToLower(logFormat) {
	case "apache":
		globalLogFormat = "{client_ip} - - {time:%d/%m/%Y:%H:%M:%S} {method} {path} {status_code} -"
	case "nginx":
		globalLogFormat = "{client_ip} - {time:%d/%b/%Y:%H:%M:%S} {method} {path} {status_code} {latency:ms}"
	case "full":
		globalLogFormat = "{time} {status_code} {method} {path} {client_ip} {latency} \"{user_agent}\" {protocol} {request_size} {response_size}"
	case "random":
		globalLogFormat = generateRandomGlobalLogFormat()
	default:
		// If user supplied custom format with placeholders, use it.
		globalLogFormat = logFormat
	}
	// Print the selected global log format.
	fmt.Println("Global Log Format:", globalLogFormat)
}

// processPort reads the PORT env variable and uses processRandomInt to support "RANDOM" values.
func processPort() int {
	portStr := viper.GetString("PORT")
	port, err := processRandomInt(portStr, 1024, 65535)
	if err != nil {
		fmt.Println("invalid PORT env var", zap.Error(err))
		return 8080
	}
	return port
}
