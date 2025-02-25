package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Payload for Simulate Concurrent Flood.
type ConcurrentFloodPayload struct {
	TargetEndpoint string  `json:"target_endpoint"` // e.g., "/simple"
	RequestCount   DuckInt `json:"request_count"`   // Number of requests per interval.
	MaintainSecond DuckInt `json:"maintain_second"` // Duration of the simulation.
	Async          bool    `json:"async"`
	IntervalSecond DuckInt `json:"interval_second"` // Interval between bursts.
}

// ConcurrentFloodHandler handles POST /stress/concurrent_flood.
func ConcurrentFloodHandler(c *gin.Context) {
	var payload ConcurrentFloodPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	reqCount := int(payload.RequestCount)
	intervalSec := int(payload.IntervalSecond)
	target := payload.TargetEndpoint

	// Define a function to run the flood.
	floodFunc := func() {
		endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
		client := &http.Client{Timeout: 5 * time.Second}
		// Build the full URL: assume the target endpoint is relative; use current host.
		fullURL := fmt.Sprintf("http://%s%s", c.Request.Host, target)
		for time.Now().Before(endTime) {
			var wg sync.WaitGroup
			for i := 0; i < reqCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					// We ignore the response; errors are logged.
					if _, err := client.Get(fullURL); err != nil {
						logger.Error("concurrent flood request failed", zap.Error(err))
					}
				}()
			}
			wg.Wait()
			time.Sleep(time.Duration(intervalSec) * time.Second)
		}
		logger.Info("Concurrent flood simulation completed", zap.Int("duration_sec", maintainSec))
	}

	if payload.Async {
		go floodFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "concurrent flood simulation started",
			"target_endpoint": target,
			"request_count":   reqCount,
			"maintain_second": maintainSec,
			"interval_second": intervalSec,
		})
	} else {
		floodFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "concurrent flood simulation completed",
			"target_endpoint": target,
			"request_count":   reqCount,
			"maintain_second": maintainSec,
			"interval_second": intervalSec,
		})
	}
}

// Payload for Simulate Downtime.
type DowntimePayload struct {
	DowntimeSecond DuckInt `json:"downtime_second"`
	Async          bool    `json:"async"`
}

// Global variable to control downtime.
var (
	downtimeActive bool
	downtimeMutex  sync.Mutex
)

// DowntimeHandler handles POST /stress/downtime.
func DowntimeHandler(c *gin.Context) {
	var payload DowntimePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	downtimeSec := int(payload.DowntimeSecond)

	// Activate downtime.
	downtimeMutex.Lock()
	downtimeActive = true
	downtimeMutex.Unlock()
	logger.Info("Downtime simulation started", zap.Int("downtime_sec", downtimeSec))

	resetFunc := func() {
		time.Sleep(time.Duration(downtimeSec) * time.Second)
		downtimeMutex.Lock()
		downtimeActive = false
		downtimeMutex.Unlock()
		logger.Info("Downtime simulation ended")
	}

	if payload.Async {
		go resetFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "downtime simulation started",
			"downtime_second": downtimeSec,
		})
	} else {
		resetFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "downtime simulation completed",
			"downtime_second": downtimeSec,
		})
	}
}

// DowntimeMiddleware intercepts requests when downtime is active.
func DowntimeMiddleware(c *gin.Context) {
	downtimeMutex.Lock()
	active := downtimeActive
	downtimeMutex.Unlock()
	if active {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"error":        "SERVICE_DOWN",
			"message":      "Service is temporarily unavailable",
			"requested_at": time.Now().UTC().Format(time.RFC3339Nano),
		})
		return
	}
	c.Next()
}

// Payload for Simulate External API Calls.
type ThirdPartyPayload struct {
	TargetURL      string  `json:"target_url"`
	MaintainSecond DuckInt `json:"maintain_second"`
	Async          bool    `json:"async"`
	CallRate       DuckInt `json:"call_rate"`       // Number of calls per interval.
	IntervalSecond DuckInt `json:"interval_second"` // Interval between bursts.
	SimulateErrors bool    `json:"simulate_errors"`
}

// ThirdPartyHandler handles POST /stress/third_party.
func ThirdPartyHandler(c *gin.Context) {
	var payload ThirdPartyPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	callRate := int(payload.CallRate)
	intervalSec := int(payload.IntervalSecond)
	targetURL := payload.TargetURL
	simErr := payload.SimulateErrors

	floodFunc := func() {
		endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
		client := &http.Client{Timeout: 5 * time.Second}
		for time.Now().Before(endTime) {
			var wg sync.WaitGroup
			for i := 0; i < callRate; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					// If simulate_errors is enabled, randomly decide to inject an error.
					if simErr && rand.Float64() < 0.2 {
						logger.Error("Simulated third-party call error")
						return
					}
					if _, err := client.Get(targetURL); err != nil {
						logger.Error("Third-party API call failed", zap.Error(err))
					}
				}()
			}
			wg.Wait()
			time.Sleep(time.Duration(intervalSec) * time.Second)
		}
		logger.Info("Third-party API call simulation completed", zap.Int("duration_sec", maintainSec))
	}

	if payload.Async {
		go floodFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "third-party API call simulation started",
			"target_url":      targetURL,
			"maintain_second": maintainSec,
			"call_rate":       callRate,
			"interval_second": intervalSec,
			"simulate_errors": simErr,
		})
	} else {
		floodFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "third-party API call simulation completed",
			"target_url":      targetURL,
			"maintain_second": maintainSec,
			"call_rate":       callRate,
			"interval_second": intervalSec,
			"simulate_errors": simErr,
		})
	}
}

// Payload for Simulate DDoS Attack.
type DDoSPayload struct {
	TargetEndpoint  string  `json:"target_endpoint"`
	AttackIntensity DuckInt `json:"attack_intensity"` // Number of requests per interval.
	MaintainSecond  DuckInt `json:"maintain_second"`
	Async           bool    `json:"async"`
	IntervalSecond  DuckInt `json:"interval_second"`
}

// DDoSHandler handles POST /stress/ddos.
func DDoSHandler(c *gin.Context) {
	var payload DDoSPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	attackIntensity := int(payload.AttackIntensity)
	intervalSec := int(payload.IntervalSecond)
	target := payload.TargetEndpoint

	ddosFunc := func() {
		endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
		client := &http.Client{Timeout: 5 * time.Second}
		fullURL := fmt.Sprintf("http://%s%s", c.Request.Host, target)
		for time.Now().Before(endTime) {
			var wg sync.WaitGroup
			for i := 0; i < attackIntensity; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if _, err := client.Get(fullURL); err != nil {
						logger.Error("DDoS attack request failed", zap.Error(err))
					}
				}()
			}
			wg.Wait()
			time.Sleep(time.Duration(intervalSec) * time.Second)
		}
		logger.Info("DDoS attack simulation completed", zap.Int("duration_sec", maintainSec))
	}

	if payload.Async {
		go ddosFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":          "DDoS attack simulation started",
			"target_endpoint":  target,
			"attack_intensity": attackIntensity,
			"maintain_second":  maintainSec,
			"interval_second":  intervalSec,
		})
	} else {
		ddosFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":          "DDoS attack simulation completed",
			"target_endpoint":  target,
			"attack_intensity": attackIntensity,
			"maintain_second":  maintainSec,
			"interval_second":  intervalSec,
		})
	}
}
