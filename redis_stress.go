package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisHeavyPayload defines the payload for heavy Redis queries using a single connection.
type RedisHeavyPayload struct {
	Reads            bool    `json:"reads"`
	Writes           bool    `json:"writes"`
	MaintainSecond   DuckInt `json:"maintain_second"`
	Async            bool    `json:"async"`
	QueryPerInterval DuckInt `json:"query_per_interval"`
	IntervalSecond   DuckInt `json:"interval_second"`
}

// RedisMultiHeavyPayload defines the payload for heavy Redis queries using multiple connections.
type RedisMultiHeavyPayload struct {
	Reads            bool    `json:"reads"`
	Writes           bool    `json:"writes"`
	MaintainSecond   DuckInt `json:"maintain_second"`
	Async            bool    `json:"async"`
	ConnectionCounts DuckInt `json:"connection_counts"`
	QueryPerInterval DuckInt `json:"query_per_interval"`
	IntervalSecond   DuckInt `json:"interval_second"`
}

// RedisConnectionPayload defines the payload for simulating heavy Redis connection load.
type RedisConnectionPayload struct {
	MaintainSecond      DuckInt `json:"maintain_second"`
	Async               bool    `json:"async"`
	ConnectionCounts    DuckInt `json:"connection_counts"`
	IncreasePerInterval DuckInt `json:"increase_per_interval"`
	IntervalSecond      DuckInt `json:"interval_second"`
}

// getRedisClient creates and returns a new Redis client using configuration from GetRedisConfig.
func getRedisClient() (*redis.Client, error) {
	cfg, err := GetRedisConfig()
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	options := &redis.Options{
		Addr: addr,
	}
	if cfg.TLSEnabled {
		options.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := redis.NewClient(options)
	// Use a background context for simplicity.
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

// RedisHeavyHandler handles POST /redis/heavy.
// It performs read/write commands on a single Redis connection for the specified duration.
func RedisHeavyHandler(c *gin.Context) {
	var payload RedisHeavyPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	queryPerInterval := int(payload.QueryPerInterval)
	intervalSec := int(payload.IntervalSecond)

	client, err := getRedisClient()
	if err != nil {
		ErrorJSON(c, 500, "REDIS_ERROR", err.Error())
		return
	}
	ctx := context.Background()

	stressFunc := func() {
		endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
		for time.Now().Before(endTime) {
			for i := 0; i < queryPerInterval; i++ {
				if payload.Reads {
					_, err := client.Get(ctx, "stress_key").Result()
					if err != nil && err != redis.Nil {
						logger.Error("Redis heavy read failed", zap.Error(err))
					}
				}
				if payload.Writes {
					if err := client.Set(ctx, "stress_key", "stress", 0).Err(); err != nil {
						logger.Error("Redis heavy write failed", zap.Error(err))
					}
				}
			}
			time.Sleep(time.Duration(intervalSec) * time.Second)
		}
		client.Close()
		logger.Info("Redis heavy query (single connection) completed", zap.Int("duration_sec", maintainSec))
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":            "Redis heavy query (single connection) started",
			"maintain_second":    maintainSec,
			"query_per_interval": queryPerInterval,
			"interval_second":    intervalSec,
		})
	} else {
		stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":            "Redis heavy query (single connection) completed",
			"maintain_second":    maintainSec,
			"query_per_interval": queryPerInterval,
			"interval_second":    intervalSec,
		})
	}
}

// RedisMultiHeavyHandler handles POST /redis/multi_heavy.
// It spawns multiple concurrent connections, each performing queries for the specified duration.
func RedisMultiHeavyHandler(c *gin.Context) {
	var payload RedisMultiHeavyPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	queryPerInterval := int(payload.QueryPerInterval)
	intervalSec := int(payload.IntervalSecond)
	connectionCounts := int(payload.ConnectionCounts)

	stressFunc := func() {
		var wg sync.WaitGroup
		for i := 0; i < connectionCounts; i++ {
			wg.Add(1)
			go func(connNum int) {
				defer wg.Done()
				client, err := getRedisClient()
				if err != nil {
					logger.Error("Redis multi heavy connection failed", zap.Int("conn", connNum), zap.Error(err))
					return
				}
				ctx := context.Background()
				endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
				for time.Now().Before(endTime) {
					for j := 0; j < queryPerInterval; j++ {
						if payload.Reads {
							_, err := client.Get(ctx, "stress_key").Result()
							if err != nil && err != redis.Nil {
								logger.Error("Redis multi heavy read failed", zap.Int("conn", connNum), zap.Error(err))
							}
						}
						if payload.Writes {
							if err := client.Set(ctx, "stress_key", "stress", 0).Err(); err != nil {
								logger.Error("Redis multi heavy write failed", zap.Int("conn", connNum), zap.Error(err))
							}
						}
					}
					time.Sleep(time.Duration(intervalSec) * time.Second)
				}
				client.Close()
			}(i)
		}
		wg.Wait()
		logger.Info("Redis multi heavy query completed", zap.Int("connections", connectionCounts))
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":            "Redis multi heavy query started",
			"maintain_second":    maintainSec,
			"query_per_interval": queryPerInterval,
			"interval_second":    intervalSec,
			"connection_counts":  connectionCounts,
		})
	} else {
		stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":            "Redis multi heavy query completed",
			"maintain_second":    maintainSec,
			"query_per_interval": queryPerInterval,
			"interval_second":    intervalSec,
			"connection_counts":  connectionCounts,
		})
	}
}

// RedisConnectionHandler handles POST /redis/connection.
// It gradually opens multiple Redis connections until reaching the target connection_counts
// and maintains them open for the specified duration.
func RedisConnectionHandler(c *gin.Context) {
	var payload RedisConnectionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	connectionCounts := int(payload.ConnectionCounts)
	increasePerInterval := int(payload.IncreasePerInterval)
	intervalSec := int(payload.IntervalSecond)

	stressFunc := func() {
		var clients []*redis.Client
		var mu sync.Mutex
		endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
		currentCount := 0
		ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
		defer ticker.Stop()

	Loop:
		for {
			select {
			case <-ticker.C:
				for i := 0; i < increasePerInterval && currentCount < connectionCounts; i++ {
					client, err := getRedisClient()
					if err != nil {
						logger.Error("Redis connection stress open failed", zap.Error(err))
						continue
					}
					mu.Lock()
					clients = append(clients, client)
					currentCount++
					mu.Unlock()
				}
				if currentCount >= connectionCounts {
					break Loop
				}
				if time.Now().After(endTime) {
					break Loop
				}
			default:
				if time.Now().After(endTime) {
					break Loop
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
		remaining := time.Until(endTime)
		if remaining > 0 {
			time.Sleep(remaining)
		}
		mu.Lock()
		for _, client := range clients {
			client.Close()
		}
		mu.Unlock()
		logger.Info("Redis connection stress completed", zap.Int("connections", currentCount))
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":               "Redis connection stress started",
			"maintain_second":       maintainSec,
			"connection_counts":     connectionCounts,
			"increase_per_interval": increasePerInterval,
			"interval_second":       intervalSec,
		})
	} else {
		stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":               "Redis connection stress completed",
			"maintain_second":       maintainSec,
			"connection_counts":     connectionCounts,
			"increase_per_interval": increasePerInterval,
			"interval_second":       intervalSec,
		})
	}
}
