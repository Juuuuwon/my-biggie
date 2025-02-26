package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

// Payload for heavy MySQL query on a single connection.
type MySQLHeavyPayload struct {
	Reads            bool    `json:"reads"`
	Writes           bool    `json:"writes"`
	MaintainSecond   DuckInt `json:"maintain_second"`
	Async            bool    `json:"async"`
	QueryPerInterval DuckInt `json:"query_per_interval"`
	IntervalSecond   DuckInt `json:"interval_second"`
}

// Payload for heavy MySQL query on multiple connections.
type MySQLMultiHeavyPayload struct {
	Reads            bool    `json:"reads"`
	Writes           bool    `json:"writes"`
	MaintainSecond   DuckInt `json:"maintain_second"`
	Async            bool    `json:"async"`
	ConnectionCounts DuckInt `json:"connection_counts"`
	QueryPerInterval DuckInt `json:"query_per_interval"`
	IntervalSecond   DuckInt `json:"interval_second"`
}

// Payload for heavy MySQL connection load.
type MySQLConnectionPayload struct {
	MaintainSecond      DuckInt `json:"maintain_second"`
	Async               bool    `json:"async"`
	ConnectionCounts    DuckInt `json:"connection_counts"`
	IncreasePerInterval DuckInt `json:"increase_per_interval"`
	IntervalSecond      DuckInt `json:"interval_second"`
}

// MySQLHeavyHandler handles POST /mysql/heavy.
// It opens a single connection and repeatedly performs read and/or write queries.
func MySQLHeavyHandler(c *gin.Context) {
	var payload MySQLHeavyPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	queryPerInterval := int(payload.QueryPerInterval)
	intervalSec := int(payload.IntervalSecond)

	cfg, err := GetMySQLConfig()
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, "CONFIG_ERROR", err.Error())
		return
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if err = db.Ping(); err != nil {
		ErrorJSON(c, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	stressFunc := func() {
		endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
		for time.Now().Before(endTime) {
			for i := 0; i < queryPerInterval; i++ {
				if payload.Reads {
					if _, err := db.Query("SELECT 1"); err != nil {
						fmt.Println("MySQL heavy read query failed", zap.Error(err))
					}
				}
				if payload.Writes {
					// Assumes table "biggie_test_table" exists.
					if _, err := db.Exec("INSERT INTO biggie_test_table(value) VALUES('stress')"); err != nil {
						fmt.Println("MySQL heavy write query failed", zap.Error(err))
					}
				}
			}
			time.Sleep(time.Duration(intervalSec) * time.Second)
		}
		db.Close()
		fmt.Println("MySQL heavy query (single connection) completed", zap.Int("duration_sec", maintainSec))
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":            "MySQL heavy query (single connection) started",
			"maintain_second":    maintainSec,
			"query_per_interval": queryPerInterval,
			"interval_second":    intervalSec,
		})
	} else {
		stressFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":            "MySQL heavy query (single connection) completed",
			"maintain_second":    maintainSec,
			"query_per_interval": queryPerInterval,
			"interval_second":    intervalSec,
		})
	}
}

// MySQLMultiHeavyHandler handles POST /mysql/multi_heavy.
// It spawns multiple concurrent connections, each performing queries for the specified duration.
func MySQLMultiHeavyHandler(c *gin.Context) {
	var payload MySQLMultiHeavyPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	queryPerInterval := int(payload.QueryPerInterval)
	intervalSec := int(payload.IntervalSecond)
	connectionCounts := int(payload.ConnectionCounts)

	cfg, err := GetMySQLConfig()
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, "CONFIG_ERROR", err.Error())
		return
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	stressFunc := func() {
		var wg sync.WaitGroup
		for i := 0; i < connectionCounts; i++ {
			wg.Add(1)
			go func(connNum int) {
				defer wg.Done()
				db, err := sql.Open("mysql", dsn)
				if err != nil {
					fmt.Println("MySQL multi heavy connection open failed", zap.Int("conn", connNum), zap.Error(err))
					return
				}
				defer db.Close()
				if err = db.Ping(); err != nil {
					fmt.Println("MySQL multi heavy ping failed", zap.Int("conn", connNum), zap.Error(err))
					return
				}
				endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
				for time.Now().Before(endTime) {
					for j := 0; j < queryPerInterval; j++ {
						if payload.Reads {
							if _, err := db.Query("SELECT 1"); err != nil {
								fmt.Println("MySQL multi heavy read query failed", zap.Int("conn", connNum), zap.Error(err))
							}
						}
						if payload.Writes {
							if _, err := db.Exec("INSERT INTO biggie_test_table(value) VALUES('stress')"); err != nil {
								fmt.Println("MySQL multi heavy write query failed", zap.Int("conn", connNum), zap.Error(err))
							}
						}
					}
					time.Sleep(time.Duration(intervalSec) * time.Second)
				}
			}(i)
		}
		wg.Wait()
		fmt.Println("MySQL multi heavy query completed", zap.Int("connections", connectionCounts))
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":            "MySQL multi heavy query started",
			"maintain_second":    maintainSec,
			"query_per_interval": queryPerInterval,
			"interval_second":    intervalSec,
			"connection_counts":  connectionCounts,
		})
	} else {
		stressFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":            "MySQL multi heavy query completed",
			"maintain_second":    maintainSec,
			"query_per_interval": queryPerInterval,
			"interval_second":    intervalSec,
			"connection_counts":  connectionCounts,
		})
	}
}

// MySQLConnectionHandler handles POST /mysql/connection.
// It gradually establishes multiple MySQL connections over the specified duration.
func MySQLConnectionHandler(c *gin.Context) {
	var payload MySQLConnectionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	connectionCounts := int(payload.ConnectionCounts)
	increasePerInterval := int(payload.IncreasePerInterval)
	intervalSec := int(payload.IntervalSecond)

	cfg, err := GetMySQLConfig()
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, "CONFIG_ERROR", err.Error())
		return
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	stressFunc := func() {
		var connections []*sql.DB
		var mu sync.Mutex
		endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
		currentCount := 0
		ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
		defer ticker.Stop()

		// Gradually open connections.
	Loop:
		for {
			select {
			case <-ticker.C:
				for i := 0; i < increasePerInterval && currentCount < connectionCounts; i++ {
					db, err := sql.Open("mysql", dsn)
					if err != nil {
						fmt.Println("MySQL connection stress open failed", zap.Error(err))
						continue
					}
					if err = db.Ping(); err != nil {
						fmt.Println("MySQL connection stress ping failed", zap.Error(err))
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
		// Maintain connections until endTime.
		remaining := time.Until(endTime)
		if remaining > 0 {
			time.Sleep(remaining)
		}
		// Close all connections.
		mu.Lock()
		for _, db := range connections {
			db.Close()
		}
		mu.Unlock()
		fmt.Println("MySQL connection stress completed", zap.Int("connections", currentCount))
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":               "MySQL connection stress started",
			"maintain_second":       maintainSec,
			"connection_counts":     connectionCounts,
			"increase_per_interval": increasePerInterval,
			"interval_second":       intervalSec,
		})
	} else {
		stressFunc()
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":               "MySQL connection stress completed",
			"maintain_second":       maintainSec,
			"connection_counts":     connectionCounts,
			"increase_per_interval": increasePerInterval,
			"interval_second":       intervalSec,
		})
	}
}
