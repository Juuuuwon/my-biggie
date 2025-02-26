package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/segmentio/kafka-go"
)

// HealthCheckHandler handles GET /healthcheck and returns "ok" as fast as possible.
func HealthCheckHandler(c *gin.Context) {
	ResponseJSON(c, http.StatusOK, gin.H{"message": "ok"})
}

// SlowHealthCheckHandler handles GET /healthcheck/slow?wait=[number].
// It waits for the specified number of seconds (or a random duration if not provided)
// before returning "ok".
func SlowHealthCheckHandler(c *gin.Context) {
	waitStr := c.Query("wait")
	var waitSec int
	var err error
	if waitStr == "" {
		// If not provided, choose a random wait between 1 and 5 seconds.
		waitSec = 1 + rand.Intn(5)
	} else {
		waitSec, err = strconv.Atoi(waitStr)
		if err != nil || waitSec < 0 {
			waitSec = 1 + rand.Intn(5)
		}
	}
	time.Sleep(time.Duration(waitSec) * time.Second)
	ResponseJSON(c, http.StatusOK, gin.H{"message": "ok"})
}

// ExternalHealthHandler handles GET /healthcheck/external.
// It tests the connection to all configured external services and returns their status.
func ExternalHealthHandler(c *gin.Context) {
	statuses := make(map[string]string)

	// Check MySQL
	if mysqlCfg, err := GetMySQLConfig(); err == nil {
		if err := checkMySQL(mysqlCfg); err != nil {
			statuses["mysql"] = fmt.Sprintf("failed: %v", err)
		} else {
			statuses["mysql"] = "ok"
		}
	} else {
		statuses["mysql"] = "not configured"
	}

	// Check PostgreSQL
	if pgCfg, err := GetPostgresConfig(); err == nil {
		if err := checkPostgres(pgCfg); err != nil {
			statuses["postgres"] = fmt.Sprintf("failed: %v", err)
		} else {
			statuses["postgres"] = "ok"
		}
	} else {
		statuses["postgres"] = "not configured"
	}

	// Check Redshift
	if rsCfg, err := GetRedshiftConfig(); err == nil {
		if err := checkRedshift(rsCfg); err != nil {
			statuses["redshift"] = fmt.Sprintf("failed: %v", err)
		} else {
			statuses["redshift"] = "ok"
		}
	} else {
		statuses["redshift"] = "not configured"
	}

	// Check Redis
	if redisCfg, err := GetRedisConfig(); err == nil {
		if err := checkRedis(redisCfg); err != nil {
			statuses["redis"] = fmt.Sprintf("failed: %v", err)
		} else {
			statuses["redis"] = "ok"
		}
	} else {
		statuses["redis"] = "not configured"
	}

	// Check Kafka
	if kafkaCfg, err := GetKafkaConfig(); err == nil {
		if err := checkKafka(kafkaCfg); err != nil {
			statuses["kafka"] = fmt.Sprintf("failed: %v", err)
		} else {
			statuses["kafka"] = "ok"
		}
	} else {
		statuses["kafka"] = "not configured"
	}

	ResponseJSON(c, http.StatusOK, statuses)
}

// RelayRequest defines the expected JSON payload for the relay API.
type RelayRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// RelayResponse defines the structure of the relay response.
type RelayResponse struct {
	StatusCode  int         `json:"status_code"`
	Headers     http.Header `json:"headers"`
	Body        string      `json:"body"`
	RequestedAt string      `json:"requested_at"`
}

// RelayHandler handles POST /healthcheck/hops.
// It sends an HTTP request to the specified URL with given method, headers, and body,
// then returns the response details.
func RelayHandler(c *gin.Context) {
	var reqPayload RelayRequest
	if err := c.ShouldBindJSON(&reqPayload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}

	// Create the new request with provided body.
	var bodyReader io.Reader
	if reqPayload.Body != "" {
		bodyReader = bytes.NewBufferString(reqPayload.Body)
	}
	req, err := http.NewRequest(reqPayload.Method, reqPayload.URL, bodyReader)
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, "REQUEST_CREATION_FAILED", err.Error())
		return
	}
	// Set provided headers.
	for key, value := range reqPayload.Headers {
		req.Header.Set(key, value)
	}

	// Create a client with a timeout.
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, "REQUEST_FAILED", err.Error())
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorJSON(c, http.StatusInternalServerError, "READ_RESPONSE_FAILED", err.Error())
		return
	}

	// Build the relay response.
	relayResp := RelayResponse{
		StatusCode:  resp.StatusCode,
		Headers:     resp.Header,
		Body:        string(respBody),
		RequestedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	ResponseJSON(c, http.StatusOK, relayResp)
}

// checkMySQL connects to MySQL using the provided configuration and pings the server.
func checkMySQL(cfg *MySQLConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}

// checkPostgres connects to PostgreSQL using the provided configuration and pings the server.
func checkPostgres(cfg *PostgresConfig) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}

// checkRedshift connects to Redshift (using pgx as driver) and pings the server.
func checkRedshift(cfg *RedshiftConfig) error {
	// Use the same DSN format as PostgreSQL.
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}

// checkRedis creates a Redis client using the provided configuration and pings the server.
func checkRedis(cfg *RedisConfig) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	options := &redis.Options{
		Addr: addr,
	}
	if cfg.TLSEnabled {
		options.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	client := redis.NewClient(options)
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return client.Ping(ctx).Err()
}

// checkKafka connects to the Kafka cluster by dialing the first server in the list.
func checkKafka(cfg *KafkaConfig) error {
	if len(cfg.Servers) == 0 {
		return fmt.Errorf("no Kafka servers provided")
	}
	conn, err := kafka.Dial("tcp", cfg.Servers[0])
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
