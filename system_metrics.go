package main

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// SystemMetricsHandler handles GET /metrics/system.
// It aggregates system metrics such as CPU load, memory usage, network throughput,
// and details of ongoing stress tests.
func SystemMetricsHandler(c *gin.Context) {
	// Dummy CPU load value (in a real implementation, you might use a library such as gopsutil).
	cpuLoad := 0.75

	// Gather memory usage metrics using runtime.ReadMemStats.
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memoryUsage := map[string]uint64{
		"alloc":       memStats.Alloc,
		"total_alloc": memStats.TotalAlloc,
		"sys":         memStats.Sys,
		"num_gc":      uint64(memStats.NumGC), // Cast uint32 to uint64.
	}

	// Dummy network throughput values (in bytes per second).
	networkThroughput := map[string]int{
		"network_in":  1024,
		"network_out": 2048,
	}

	// Gather stress test details from global variables.
	stressTests := map[string]interface{}{
		"error_injection_rate":   activeErrorRate,
		"network_latency_ms":     activeLatencyMs,
		"packet_loss_percentage": activePacketLoss,
	}

	// Include downtime status (accessed via mutex).
	downtimeMutex.Lock()
	stressTests["downtime_active"] = downtimeActive
	downtimeMutex.Unlock()

	// Aggregate all metrics.
	metrics := map[string]interface{}{
		"cpu_load":           cpuLoad,
		"memory_usage":       memoryUsage,
		"network_throughput": networkThroughput,
		"stress_tests":       stressTests,
		"requested_at":       time.Now().UTC().Format(time.RFC3339Nano),
	}

	c.JSON(http.StatusOK, metrics)
}
