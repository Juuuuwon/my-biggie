package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// processRandomInt checks if the provided string uses the RANDOM syntax for integers.
// If the value is exactly "RANDOM", it returns a random integer between defaultStart and defaultEnd.
// If it follows "RANDOM:<start>:<end>", it returns a random integer in that range.
// Otherwise, it attempts to parse the value as an integer.
func processRandomInt(value string, defaultStart, defaultEnd int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "RANDOM" {
		return rand.Intn(defaultEnd-defaultStart) + defaultStart, nil
	}
	if strings.HasPrefix(value, "RANDOM:") {
		parts := strings.Split(value, ":")
		if len(parts) != 3 {
			return 0, errors.New("invalid RANDOM syntax for integer")
		}
		start, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, err
		}
		end, err := strconv.Atoi(parts[2])
		if err != nil {
			return 0, err
		}
		if start >= end {
			return 0, errors.New("invalid RANDOM range for integer: start must be less than end")
		}
		return rand.Intn(end-start) + start, nil
	}
	return strconv.Atoi(value)
}

// DuckInt is a custom type that supports duck-typing for JSON numeric fields.
// It accepts either a number or a string value (which may be "RANDOM" or "RANDOM:<start>:<end>").
type DuckInt int

// UnmarshalJSON implements json.Unmarshaler for DuckInt.
func (d *DuckInt) UnmarshalJSON(b []byte) error {
	// Try unmarshaling as an integer.
	var n int
	if err := json.Unmarshal(b, &n); err == nil {
		*d = DuckInt(n)
		return nil
	}

	// Otherwise, unmarshal as a string.
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	s = strings.TrimSpace(s)
	v, err := processRandomValue(s)
	if err != nil {
		return err
	}
	// Expect an integer result.
	switch val := v.(type) {
	case int:
		*d = DuckInt(val)
		return nil
	case string:
		n, err := strconv.Atoi(val)
		if err != nil {
			return err
		}
		*d = DuckInt(n)
		return nil
	default:
		return errors.New("unexpected type for DuckInt")
	}
}

// DuckFloat is a custom type that supports duck-typing for JSON float fields.
// It accepts either a float value or a string (which may be "RANDOM" or "RANDOM:<start>:<end>").
type DuckFloat float64

// UnmarshalJSON implements json.Unmarshaler for DuckFloat.
func (d *DuckFloat) UnmarshalJSON(b []byte) error {
	// Try unmarshaling as float64.
	var f float64
	if err := json.Unmarshal(b, &f); err == nil {
		*d = DuckFloat(f)
		return nil
	}
	// Otherwise, unmarshal as string.
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	s = strings.TrimSpace(s)
	if s == "RANDOM" {
		*d = DuckFloat(rand.Float64())
		return nil
	}
	if strings.HasPrefix(s, "RANDOM:") {
		parts := strings.Split(s, ":")
		if len(parts) != 3 {
			return errors.New("invalid RANDOM syntax for DuckFloat")
		}
		start, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return err
		}
		end, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return err
		}
		if start >= end {
			return errors.New("invalid RANDOM range for DuckFloat")
		}
		*d = DuckFloat(start + rand.Float64()*(end-start))
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*d = DuckFloat(f)
	return nil
}

// processRandomValue checks if the provided string uses the RANDOM syntax.
// If the value is exactly "RANDOM", it returns a generated random string.
// If it follows "RANDOM:<start>:<end>", it returns a random integer within that range.
func processRandomValue(value string) (interface{}, error) {
	if value == "RANDOM" {
		return "randomValue-" + strconv.Itoa(rand.Intn(10000)), nil
	}
	if strings.HasPrefix(value, "RANDOM:") {
		parts := strings.Split(value, ":")
		if len(parts) != 3 {
			return nil, errors.New("invalid RANDOM syntax")
		}
		start, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		end, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, err
		}
		if start >= end {
			return nil, errors.New("invalid RANDOM range: start must be less than end")
		}
		return rand.Intn(end-start) + start, nil
	}
	return value, nil
}

// ResponseJSON writes a JSON response with an automatically added "requested_at" timestamp.
func ResponseJSON(c *gin.Context, status int, payload interface{}) {
	response := gin.H{
		"requested_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	if payloadMap, ok := payload.(gin.H); ok {
		for k, v := range payloadMap {
			response[k] = v
		}
	} else {
		response["data"] = payload
	}
	c.JSON(status, response)
}

// ErrorJSON sends a standardized JSON error response.
func ErrorJSON(c *gin.Context, status int, errorType, message string) {
	errResp := gin.H{
		"error":        strings.ToUpper(errorType),
		"message":      strings.ToLower(message),
		"request":      getRequestDetails(c),
		"requested_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	c.JSON(status, errResp)
}

// getRequestDetails extracts basic details from the incoming HTTP request.
func getRequestDetails(c *gin.Context) gin.H {
	details := gin.H{
		"method": c.Request.Method,
		"ip":     c.ClientIP(),
		"query":  c.Request.URL.Query(),
	}
	cookies := make(map[string]string)
	for _, cookie := range c.Request.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	details["cookies"] = cookies
	if rawBody, exists := c.Get("rawBody"); exists {
		bodyStr := rawBody.(string)
		details["body"] = gin.H{
			"length":  len(bodyStr),
			"payload": bodyStr,
		}
	}
	return details
}

// RequestBodyMiddleware reads the raw request body and stores it in the Gin context.
func RequestBodyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				c.Set("rawBody", string(bodyBytes))
			}
		}
		c.Next()
	}
}
