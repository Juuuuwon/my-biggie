package main

import (
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorInjectionPayload defines the JSON payload for the error injection API.
type ErrorInjectionPayload struct {
	ErrorRate      DuckFloat `json:"error_rate"`      // Supports duck-typing for error rate (e.g., "RANDOM:0.05:0.15")
	MaintainSecond DuckInt   `json:"maintain_second"` // Supports RANDOM syntax via DuckInt.
	Async          bool      `json:"async"`
}

// CrashSimulationPayload defines the JSON payload for the crash simulation API.
type CrashSimulationPayload struct {
	MaintainSecond DuckInt `json:"maintain_second"`
	Async          bool    `json:"async"`
}

// Global variables to control error injection.
var (
	activeErrorRate      float64   = 0.0
	errorInjectionExpiry time.Time = time.Now()
)

// ErrorInjectionHandler handles POST /stress/error_injection.
// It sets a global error injection rate for the specified duration.
func ErrorInjectionHandler(c *gin.Context) {
	var payload ErrorInjectionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	durationSec := int(payload.MaintainSecond)
	// Convert DuckFloat to float64.
	activeErrorRate = float64(payload.ErrorRate)
	errorInjectionExpiry = time.Now().Add(time.Duration(durationSec) * time.Second)
	logger.Info("Error injection started",
		zap.Float64("error_rate", activeErrorRate),
		zap.Int("duration_sec", durationSec))

	resetFunc := func() {
		time.Sleep(time.Duration(durationSec) * time.Second)
		activeErrorRate = 0.0
		logger.Info("Error injection ended")
	}

	if payload.Async {
		go resetFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "error injection started",
			"error_rate":      activeErrorRate,
			"maintain_second": durationSec,
		})
	} else {
		resetFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "error injection completed",
			"error_rate":      activeErrorRate,
			"maintain_second": durationSec,
		})
	}
}

// CrashSimulationHandler handles POST /stress/crash.
// It simulates a crash by exiting the process after the specified duration.
func CrashSimulationHandler(c *gin.Context) {
	var payload CrashSimulationPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	durationSec := int(payload.MaintainSecond)
	logger.Info("Crash simulation scheduled", zap.Int("maintain_second", durationSec))

	crashFunc := func() {
		time.Sleep(time.Duration(durationSec) * time.Second)
		logger.Error("Simulated crash: exiting process")
		os.Exit(1)
	}

	if payload.Async {
		go crashFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "crash simulation started",
			"maintain_second": durationSec,
		})
	} else {
		crashFunc()
		// Will not reach here.
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "crash simulation completed",
			"maintain_second": durationSec,
		})
	}
}

// ErrorInjectionMiddleware is a global middleware that, if error injection is active,
// randomly aborts requests with an error response based on the active error rate.
func ErrorInjectionMiddleware(c *gin.Context) {
	if time.Now().Before(errorInjectionExpiry) && activeErrorRate > 0 {
		if rand.Float64() < activeErrorRate {
			ErrorJSON(c, http.StatusInternalServerError, "RANDOM_ERROR", "simulated random error injection")
			c.Abort()
			return
		}
	}
	c.Next()
}
