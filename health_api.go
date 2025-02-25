package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
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
