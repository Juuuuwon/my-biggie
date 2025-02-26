package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Global variables for network stress simulation.
var (
	networkStressMutex sync.Mutex
	activeLatencyMs    int       = 0
	latencyExpiry      time.Time = time.Now()
	activePacketLoss   int       = 0 // Percentage (0-100)
	packetLossExpiry   time.Time = time.Now()
)

// NetworkLatencyPayload defines the payload for network latency simulation.
type NetworkLatencyPayload struct {
	LatencyMs      DuckInt `json:"latency_ms"`      // Delay in milliseconds.
	MaintainSecond DuckInt `json:"maintain_second"` // Duration.
	Async          bool    `json:"async"`
}

// NetworkLatencyHandler handles POST /stress/network/latency.
func NetworkLatencyHandler(c *gin.Context) {
	var payload NetworkLatencyPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	latencyMs := int(payload.LatencyMs)
	maintainSec := int(payload.MaintainSecond)

	// Function to set latency for the specified duration.
	setLatency := func() {
		networkStressMutex.Lock()
		activeLatencyMs = latencyMs
		latencyExpiry = time.Now().Add(time.Duration(maintainSec) * time.Second)
		networkStressMutex.Unlock()
		time.Sleep(time.Duration(maintainSec) * time.Second)
		networkStressMutex.Lock()
		activeLatencyMs = 0
		networkStressMutex.Unlock()
		fmt.Println("Network latency simulation ended", zap.Int("latency_ms", latencyMs))
	}

	if payload.Async {
		go setLatency()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "network latency simulation started",
			"latency_ms":      latencyMs,
			"maintain_second": maintainSec,
		})
	} else {
		setLatency()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "network latency simulation completed",
			"latency_ms":      latencyMs,
			"maintain_second": maintainSec,
		})
	}
}

// PacketLossPayload defines the payload for packet loss simulation.
type PacketLossPayload struct {
	LossPercentage DuckInt `json:"loss_percentage"` // Percentage of dropped requests.
	MaintainSecond DuckInt `json:"maintain_second"` // Duration.
	Async          bool    `json:"async"`
}

// PacketLossHandler handles POST /stress/network/packet_loss.
func PacketLossHandler(c *gin.Context) {
	var payload PacketLossPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	lossPercentage := int(payload.LossPercentage)
	maintainSec := int(payload.MaintainSecond)

	// Function to set packet loss for the specified duration.
	setPacketLoss := func() {
		networkStressMutex.Lock()
		activePacketLoss = lossPercentage
		packetLossExpiry = time.Now().Add(time.Duration(maintainSec) * time.Second)
		networkStressMutex.Unlock()
		time.Sleep(time.Duration(maintainSec) * time.Second)
		networkStressMutex.Lock()
		activePacketLoss = 0
		networkStressMutex.Unlock()
		fmt.Println("Packet loss simulation ended", zap.Int("loss_percentage", lossPercentage))
	}

	if payload.Async {
		go setPacketLoss()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "packet loss simulation started",
			"loss_percentage": lossPercentage,
			"maintain_second": maintainSec,
		})
	} else {
		setPacketLoss()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "packet loss simulation completed",
			"loss_percentage": lossPercentage,
			"maintain_second": maintainSec,
		})
	}
}
