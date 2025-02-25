package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CPUStressPayload defines the payload for the CPU stress test.
type CPUStressPayload struct {
	CPUPercent     DuckInt `json:"cpu_percent"`
	MaintainSecond DuckInt `json:"maintain_second"`
	Async          bool    `json:"async"`
}

// MemoryStressPayload defines the payload for the memory stress test.
type MemoryStressPayload struct {
	MemoryPercent  DuckInt `json:"memory_percent"`
	MaintainSecond DuckInt `json:"maintain_second"`
	Async          bool    `json:"async"`
}

// MemoryLeakPayload defines the payload for the memory leak simulation.
type MemoryLeakPayload struct {
	LeakSizeMB     DuckInt `json:"leak_size_mb"`
	MaintainSecond DuckInt `json:"maintain_second"`
	Async          bool    `json:"async"`
}

// Global store for memory leak simulation.
var memoryLeakStore [][]byte
var memoryLeakMutex sync.Mutex

// CPUStressHandler handles POST /stress/cpu.
// It runs a busy loop in cycles to approximate the given CPU percentage.
func CPUStressHandler(c *gin.Context) {
	var payload CPUStressPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	cpuPercent := int(payload.CPUPercent)
	maintainSec := int(payload.MaintainSecond)
	if payload.Async {
		go runCPUStress(cpuPercent, maintainSec)
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":            "cpu stress started",
			"chosen_cpu_percent": cpuPercent,
			"maintain_second":    maintainSec,
		})
	} else {
		runCPUStress(cpuPercent, maintainSec)
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":            "cpu stress completed",
			"chosen_cpu_percent": cpuPercent,
			"maintain_second":    maintainSec,
		})
	}
}

func runCPUStress(cpuPercent, maintainSec int) {
	duration := time.Duration(maintainSec) * time.Second
	endTime := time.Now().Add(duration)
	// Define a cycle period (e.g., 100ms).
	cycle := 100 * time.Millisecond
	// Calculate busy and sleep durations based on the requested CPU percentage.
	busyTime := time.Duration(cpuPercent) * cycle / 100
	sleepTime := cycle - busyTime

	for time.Now().Before(endTime) {
		start := time.Now()
		// Busy loop for busyTime.
		for {
			if time.Since(start) >= busyTime {
				break
			}
		}
		time.Sleep(sleepTime)
	}
	logger.Info("CPU stress test completed",
		zap.Int("cpu_percent", cpuPercent),
		zap.Int("duration_sec", maintainSec))
}

// MemoryStressHandler handles POST /stress/memory.
// It allocates a block of memory proportional to memory_percent (assuming a 100MB baseline for 100% stress)
// and holds it for the specified duration.
func MemoryStressHandler(c *gin.Context) {
	var payload MemoryStressPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	memoryPercent := int(payload.MemoryPercent)
	maintainSec := int(payload.MaintainSecond)
	if payload.Async {
		go runMemoryStress(memoryPercent, maintainSec)
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":               "memory stress started",
			"chosen_memory_percent": memoryPercent,
			"maintain_second":       maintainSec,
		})
	} else {
		runMemoryStress(memoryPercent, maintainSec)
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":               "memory stress completed",
			"chosen_memory_percent": memoryPercent,
			"maintain_second":       maintainSec,
		})
	}
}

func runMemoryStress(memoryPercent, maintainSec int) {
	// Assume a baseline of 100MB for 100% stress.
	allocMB := memoryPercent // e.g., 30 means 30MB.
	blockSize := allocMB * 1024 * 1024
	memBlock := make([]byte, blockSize)
	// Fill the block to prevent compiler optimizations.
	for i := range memBlock {
		memBlock[i] = byte(rand.Intn(256))
	}
	// Hold the allocation for the specified duration.
	time.Sleep(time.Duration(maintainSec) * time.Second)
	logger.Info("Memory stress test completed",
		zap.Int("memory_percent", memoryPercent),
		zap.Int("duration_sec", maintainSec))
	// The allocated memory will be freed when this function returns.
}

// MemoryLeakHandler handles POST /stress/memory_leak.
// It gradually allocates memory blocks over the specified duration and stores them globally
// to simulate a memory leak.
func MemoryLeakHandler(c *gin.Context) {
	var payload MemoryLeakPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	leakSizeMB := int(payload.LeakSizeMB)
	maintainSec := int(payload.MaintainSecond)
	if payload.Async {
		go runMemoryLeak(leakSizeMB, maintainSec)
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":             "memory leak simulation started",
			"chosen_leak_size_mb": leakSizeMB,
			"maintain_second":     maintainSec,
		})
	} else {
		runMemoryLeak(leakSizeMB, maintainSec)
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":             "memory leak simulation completed",
			"chosen_leak_size_mb": leakSizeMB,
			"maintain_second":     maintainSec,
		})
	}
}

func runMemoryLeak(leakSizeMB, maintainSec int) {
	totalBytes := leakSizeMB * 1024 * 1024
	// Allocate memory in intervals; here we allocate every 500ms.
	interval := 500 * time.Millisecond
	allocations := int((time.Duration(maintainSec) * time.Second) / interval)
	if allocations <= 0 {
		allocations = 1
	}
	bytesPerAlloc := totalBytes / allocations
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	done := time.After(time.Duration(maintainSec) * time.Second)
	for {
		select {
		case <-ticker.C:
			memBlock := make([]byte, bytesPerAlloc)
			for i := range memBlock {
				memBlock[i] = byte(rand.Intn(256))
			}
			memoryLeakMutex.Lock()
			memoryLeakStore = append(memoryLeakStore, memBlock)
			memoryLeakMutex.Unlock()
		case <-done:
			logger.Info("Memory leak simulation completed", zap.Int("leak_size_mb", leakSizeMB))
			return
		}
	}
}
