package main

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/spf13/viper"
)

// MySQLConfig holds credentials for MySQL connections.
type MySQLConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Engine   string `json:"engine"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
}

// GetMySQLConfig retrieves MySQL configuration in the following order:
// 1. MYSQL_SECRET and AWS_REGION (retrieved from a secrets manager)
// 2. MYSQL_DBINFO (JSON credentials)
// 3. Individual variables: MYSQL_HOST, MYSQL_PORT, MYSQL_USERNAME, MYSQL_PASSWORD, MYSQL_DBNAME
func GetMySQLConfig() (*MySQLConfig, error) {
	// Priority 1: MYSQL_SECRET (with AWS_REGION)
	if viper.IsSet("MYSQL_SECRET") && viper.IsSet("AWS_REGION") {
		secretStr := viper.GetString("MYSQL_SECRET")
		var cfg MySQLConfig
		if err := json.Unmarshal([]byte(secretStr), &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	// Priority 2: MYSQL_DBINFO
	if viper.IsSet("MYSQL_DBINFO") {
		dbinfoStr := viper.GetString("MYSQL_DBINFO")
		var cfg MySQLConfig
		if err := json.Unmarshal([]byte(dbinfoStr), &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	// Priority 3: Individual variables
	host := viper.GetString("MYSQL_HOST")
	if host == "" {
		return nil, errors.New("MySQL configuration not found")
	}
	port, err := processRandomInt(viper.GetString("MYSQL_PORT"), 3306, 3306)
	if err != nil {
		return nil, err
	}
	cfg := &MySQLConfig{
		Username: viper.GetString("MYSQL_USERNAME"),
		Password: viper.GetString("MYSQL_PASSWORD"),
		Engine:   "mysql",
		Host:     host,
		Port:     port,
		DBName:   viper.GetString("MYSQL_DBNAME"),
	}
	// TODO: Automatically create schemas/tables for testing if needed.
	return cfg, nil
}

// PostgresConfig holds credentials for PostgreSQL connections.
type PostgresConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Engine   string `json:"engine"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
}

// GetPostgresConfig retrieves PostgreSQL configuration in the following order:
// 1. POSTGRES_SECRET and AWS_REGION
// 2. POSTGRES_DBINFO (JSON credentials)
// 3. Individual variables: POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USERNAME, POSTGRES_PASSWORD, POSTGRES_DBNAME
func GetPostgresConfig() (*PostgresConfig, error) {
	if viper.IsSet("POSTGRES_SECRET") && viper.IsSet("AWS_REGION") {
		secretStr := viper.GetString("POSTGRES_SECRET")
		var cfg PostgresConfig
		if err := json.Unmarshal([]byte(secretStr), &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	if viper.IsSet("POSTGRES_DBINFO") {
		dbinfoStr := viper.GetString("POSTGRES_DBINFO")
		var cfg PostgresConfig
		if err := json.Unmarshal([]byte(dbinfoStr), &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	host := viper.GetString("POSTGRES_HOST")
	if host == "" {
		return nil, errors.New("PostgreSQL configuration not found")
	}
	port, err := processRandomInt(viper.GetString("POSTGRES_PORT"), 5432, 5432)
	if err != nil {
		return nil, err
	}
	cfg := &PostgresConfig{
		Username: viper.GetString("POSTGRES_USERNAME"),
		Password: viper.GetString("POSTGRES_PASSWORD"),
		Engine:   "postgres",
		Host:     host,
		Port:     port,
		DBName:   viper.GetString("POSTGRES_DBNAME"),
	}
	// TODO: Automatically create schemas/tables for testing if needed.
	return cfg, nil
}

// RedshiftConfig holds credentials for Redshift connections.
type RedshiftConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Engine   string `json:"engine"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
}

// GetRedshiftConfig retrieves Redshift configuration in the following order:
// 1. REDSHIFT_SECRET and AWS_REGION
// 2. REDSHIFT_DBINFO (JSON credentials)
// 3. Individual variables: REDSHIFT_HOST, REDSHIFT_PORT, REDSHIFT_USERNAME, REDSHIFT_PASSWORD, REDSHIFT_DBNAME
func GetRedshiftConfig() (*RedshiftConfig, error) {
	if viper.IsSet("REDSHIFT_SECRET") && viper.IsSet("AWS_REGION") {
		secretStr := viper.GetString("REDSHIFT_SECRET")
		var cfg RedshiftConfig
		if err := json.Unmarshal([]byte(secretStr), &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	if viper.IsSet("REDSHIFT_DBINFO") {
		dbinfoStr := viper.GetString("REDSHIFT_DBINFO")
		var cfg RedshiftConfig
		if err := json.Unmarshal([]byte(dbinfoStr), &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	host := viper.GetString("REDSHIFT_HOST")
	if host == "" {
		return nil, errors.New("Redshift configuration not found")
	}
	port, err := processRandomInt(viper.GetString("REDSHIFT_PORT"), 5439, 5439)
	if err != nil {
		return nil, err
	}
	cfg := &RedshiftConfig{
		Username: viper.GetString("REDSHIFT_USERNAME"),
		Password: viper.GetString("REDSHIFT_PASSWORD"),
		Engine:   "redshift",
		Host:     host,
		Port:     port,
		DBName:   viper.GetString("REDSHIFT_DBNAME"),
	}
	// TODO: Automatically create schemas/tables for testing if needed.
	return cfg, nil
}

// RedisConfig holds configuration for Redis.
type RedisConfig struct {
	Host       string
	Port       int
	TLSEnabled bool
}

// GetRedisConfig retrieves Redis configuration using individual variables:
// REDIS_HOST, REDIS_PORT, and REDIS_TLS_ENABLED.
func GetRedisConfig() (*RedisConfig, error) {
	host := viper.GetString("REDIS_HOST")
	if host == "" {
		return nil, errors.New("Redis configuration not found")
	}
	port, err := processRandomInt(viper.GetString("REDIS_PORT"), 6379, 6379)
	if err != nil {
		return nil, err
	}
	tlsStr := viper.GetString("REDIS_TLS_ENABLED")
	tlsEnabled := strings.ToLower(tlsStr) == "true"
	return &RedisConfig{
		Host:       host,
		Port:       port,
		TLSEnabled: tlsEnabled,
	}, nil
}

// KafkaConfig holds configuration for Kafka.
type KafkaConfig struct {
	Servers    []string
	TLSEnabled bool
	Topic      string
}

// GetKafkaConfig retrieves Kafka configuration using individual variables:
// KAFKA_SERVERS, KAFKA_TLS_ENABLED, and KAFKA_TOPIC.
func GetKafkaConfig() (*KafkaConfig, error) {
	serversStr := viper.GetString("KAFKA_SERVERS")
	if serversStr == "" {
		return nil, errors.New("Kafka configuration not found")
	}
	servers := strings.Split(serversStr, ",")
	for i, server := range servers {
		servers[i] = strings.TrimSpace(server)
	}
	tlsStr := viper.GetString("KAFKA_TLS_ENABLED")
	tlsEnabled := strings.ToLower(tlsStr) == "true"
	topic := viper.GetString("KAFKA_TOPIC")
	if topic == "" {
		return nil, errors.New("KAFKA_TOPIC not provided")
	}
	return &KafkaConfig{
		Servers:    servers,
		TLSEnabled: tlsEnabled,
		Topic:      topic,
	}, nil
}
