package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
)

// PostgresHeavyPayload defines the payload for heavy PostgreSQL query using a single connection.
type PostgresHeavyPayload struct {
	Reads            bool    `json:"reads"`
	Writes           bool    `json:"writes"`
	MaintainSecond   DuckInt `json:"maintain_second"`
	Async            bool    `json:"async"`
	QueryPerInterval DuckInt `json:"query_per_interval"`
	IntervalSecond   DuckInt `json:"interval_second"`
}

// PostgresMultiHeavyPayload defines the payload for heavy PostgreSQL queries using multiple connections.
type PostgresMultiHeavyPayload struct {
	Reads            bool    `json:"reads"`
	Writes           bool    `json:"writes"`
	MaintainSecond   DuckInt `json:"maintain_second"`
	Async            bool    `json:"async"`
	ConnectionCounts DuckInt `json:"connection_counts"`
	QueryPerInterval DuckInt `json:"query_per_interval"`
	IntervalSecond   DuckInt `json:"interval_second"`
}

// PostgresConnectionPayload defines the payload for simulating heavy PostgreSQL connection load.
type PostgresConnectionPayload struct {
	MaintainSecond      DuckInt `json:"maintain_second"`
	Async               bool    `json:"async"`
	ConnectionCounts    DuckInt `json:"connection_counts"`
	IncreasePerInterval DuckInt `json:"increase_per_interval"`
	IntervalSecond      DuckInt `json:"interval_second"`
}

// PostgresHeavyHandler handles POST /postgres/heavy.
// It opens a single connection and repeatedly executes read/write queries for the specified duration.
func PostgresHeavyHandler(c *gin.Context) {
	var payload PostgresHeavyPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	queryPerInterval := int(payload.QueryPerInterval)
	intervalSec := int(payload.IntervalSecond)

	cfg, err := GetPostgresConfig()
	if err != nil {
		ErrorJSON(c, 500, "CONFIG_ERROR", err.Error())
		return
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		ErrorJSON(c, 500, "DB_ERROR", err.Error())
		return
	}
	if err = db.Ping(); err != nil {
		ErrorJSON(c, 500, "DB_ERROR", err.Error())
		return
	}

	stressFunc := func() {
		endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
		for time.Now().Before(endTime) {
			for i := 0; i < queryPerInterval; i++ {
				if payload.Reads {
					if _, err := db.Query("SELECT 1"); err != nil {
						log("Postgres heavy read query failed", zap.Error(err))
					}
				}
				if payload.Writes {
					if _, err := db.Exec("INSERT INTO biggie_test_table(value) VALUES('stress')"); err != nil {
						log("Postgres heavy write query failed", zap.Error(err))
					}
				}
			}
			time.Sleep(time.Duration(intervalSec) * time.Second)
		}
		db.Close()
		log("Postgres heavy query (single connection) completed", zap.Int("duration_sec", maintainSec))
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":            "Postgres heavy query (single connection) started",
			"maintain_second":    maintainSec,
			"query_per_interval": queryPerInterval,
			"interval_second":    intervalSec,
		})
	} else {
		stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":            "Postgres heavy query (single connection) completed",
			"maintain_second":    maintainSec,
			"query_per_interval": queryPerInterval,
			"interval_second":    intervalSec,
		})
	}
}

// PostgresMultiHeavyHandler handles POST /postgres/multi_heavy.
// It spawns multiple concurrent connections (as specified by connection_counts)
// with each connection executing queries for the specified duration.
func PostgresMultiHeavyHandler(c *gin.Context) {
	var payload PostgresMultiHeavyPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	queryPerInterval := int(payload.QueryPerInterval)
	intervalSec := int(payload.IntervalSecond)
	connectionCounts := int(payload.ConnectionCounts)

	cfg, err := GetPostgresConfig()
	if err != nil {
		ErrorJSON(c, 500, "CONFIG_ERROR", err.Error())
		return
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	stressFunc := func() {
		var wg sync.WaitGroup
		for i := 0; i < connectionCounts; i++ {
			wg.Add(1)
			go func(connNum int) {
				defer wg.Done()
				db, err := sql.Open("pgx", dsn)
				if err != nil {
					log("Postgres multi heavy connection open failed", zap.Int("conn", connNum), zap.Error(err))
					return
				}
				defer db.Close()
				if err = db.Ping(); err != nil {
					log("Postgres multi heavy ping failed", zap.Int("conn", connNum), zap.Error(err))
					return
				}
				endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
				for time.Now().Before(endTime) {
					for j := 0; j < queryPerInterval; j++ {
						if payload.Reads {
							if _, err := db.Query("SELECT 1"); err != nil {
								log("Postgres multi heavy read query failed", zap.Int("conn", connNum), zap.Error(err))
							}
						}
						if payload.Writes {
							if _, err := db.Exec("INSERT INTO biggie_test_table(value) VALUES('stress')"); err != nil {
								log("Postgres multi heavy write query failed", zap.Int("conn", connNum), zap.Error(err))
							}
						}
					}
					time.Sleep(time.Duration(intervalSec) * time.Second)
				}
			}(i)
		}
		wg.Wait()
		log("Postgres multi heavy query completed", zap.Int("connections", connectionCounts))
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":            "Postgres multi heavy query started",
			"maintain_second":    maintainSec,
			"query_per_interval": queryPerInterval,
			"interval_second":    intervalSec,
			"connection_counts":  connectionCounts,
		})
	} else {
		stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":            "Postgres multi heavy query completed",
			"maintain_second":    maintainSec,
			"query_per_interval": queryPerInterval,
			"interval_second":    intervalSec,
			"connection_counts":  connectionCounts,
		})
	}
}

// PostgresConnectionHandler handles POST /postgres/connection.
// It simulates heavy connection load by gradually establishing multiple connections.
func PostgresConnectionHandler(c *gin.Context) {
	var payload PostgresConnectionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	connectionCounts := int(payload.ConnectionCounts)
	increasePerInterval := int(payload.IncreasePerInterval)
	intervalSec := int(payload.IntervalSecond)

	cfg, err := GetPostgresConfig()
	if err != nil {
		ErrorJSON(c, 500, "CONFIG_ERROR", err.Error())
		return
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	stressFunc := func() {
		var connections []*sql.DB
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
					db, err := sql.Open("pgx", dsn)
					if err != nil {
						log("Postgres connection stress open failed", zap.Error(err))
						continue
					}
					if err = db.Ping(); err != nil {
						log("Postgres connection stress ping failed", zap.Error(err))
						db.Close()
						continue
					}
					mu.Lock()
					connections = append(connections, db)
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
		for _, db := range connections {
			db.Close()
		}
		mu.Unlock()
		log("Postgres connection stress completed", zap.Int("connections", currentCount))
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":               "Postgres connection stress started",
			"maintain_second":       maintainSec,
			"connection_counts":     connectionCounts,
			"increase_per_interval": increasePerInterval,
			"interval_second":       intervalSec,
		})
	} else {
		stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":               "Postgres connection stress completed",
			"maintain_second":       maintainSec,
			"connection_counts":     connectionCounts,
			"increase_per_interval": increasePerInterval,
			"interval_second":       intervalSec,
		})
	}
}
