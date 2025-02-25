package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger and configuration.
	initLogger()
	defer logger.Sync()
	initConfig()

	// Simulate startup delay based on STARTUP_DELAY_SECOND env variable.
	startupDelay, err := processRandomInt(viper.GetString("STARTUP_DELAY_SECOND"), 1, 5) // default delay range 1-5 seconds
	if err != nil {
		logger.Warn("invalid STARTUP_DELAY_SECOND, defaulting to no delay", zap.Error(err))
	} else {
		logger.Info("startup delay", zap.Int("delay", startupDelay))
		time.Sleep(time.Duration(startupDelay) * time.Second)
	}

	gin.SetMode(gin.ReleaseMode)

	// Create a Gin router with custom middleware.
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(ZapLoggerMiddleware())
	router.Use(RequestBodyMiddleware())
	router.Use(DowntimeMiddleware)
	router.Use(NetworkStressMiddleware)
	router.Use(ErrorInjectionMiddleware)

	router.GET("/", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "text/html")
		ctx.String(200, frontendCode)
	})

	router.GET("/simple", SimpleHandler)
	router.GET("/simple/foo", FooHandler)
	router.POST("/simple/bar", BarHandler)
	router.GET("/simple/color", ColorHandler)
	router.GET("/simple/large", LargeHandler)

	router.GET("/healthcheck", HealthCheckHandler)
	router.GET("/healthcheck/slow", SlowHealthCheckHandler)
	router.GET("/healthcheck/external", ExternalHealthHandler)

	router.GET("/metadata/all", MetadataAllHandler)
	router.GET("/metadata/revision_color", RevisionColorHandler)

	router.POST("/stress/cpu", CPUStressHandler)
	router.POST("/stress/memory", MemoryStressHandler)
	router.POST("/stress/memory_leak", MemoryLeakHandler)

	router.POST("/stress/filesystem/write", FileWriteHandler)
	router.POST("/stress/filesystem/read", FileReadHandler)
	router.POST("/stress/network/latency", NetworkLatencyHandler)
	router.POST("/stress/network/packet_loss", PacketLossHandler)

	router.POST("/mysql/heavy", MySQLHeavyHandler)
	router.POST("/mysql/multi_heavy", MySQLMultiHeavyHandler)
	router.POST("/mysql/connection", MySQLConnectionHandler)

	router.POST("/postgres/heavy", PostgresHeavyHandler)
	router.POST("/postgres/multi_heavy", PostgresMultiHeavyHandler)
	router.POST("/postgres/connection", PostgresConnectionHandler)

	router.POST("/redshift/heavy", RedshiftHeavyHandler)
	router.POST("/redshift/multi_heavy", RedshiftMultiHeavyHandler)
	router.POST("/redshift/connection", RedshiftConnectionHandler)

	router.POST("/redis/heavy", RedisHeavyHandler)
	router.POST("/redis/multi_heavy", RedisMultiHeavyHandler)
	router.POST("/redis/connection", RedisConnectionHandler)

	router.POST("/kafka/heavy", KafkaHeavyHandler)
	router.POST("/kafka/multi_heavy", KafkaMultiHeavyHandler)
	router.POST("/kafka/connection", KafkaConnectionHandler)

	router.POST("/stress/error_injection", ErrorInjectionHandler)
	router.POST("/stress/crash", CrashSimulationHandler)

	router.POST("/stress/concurrent_flood", ConcurrentFloodHandler)
	router.POST("/stress/downtime", DowntimeHandler)
	router.POST("/stress/third_party", ThirdPartyHandler)
	router.POST("/stress/ddos", DDoSHandler)

	router.GET("/metrics/system", SystemMetricsHandler)

	// Determine port using environment variable (with RANDOM support).
	port := processPort()
	logger.Info("starting server", zap.Int("port", port))
	router.Run(":" + intToString(port))
}

// intToString converts an int to a string.
func intToString(i int) string {
	return fmt.Sprintf("%d", i)
}

// NetworkStressMiddleware applies active network latency and packet loss simulation.
func NetworkStressMiddleware(c *gin.Context) {
	// Check if network latency is active.
	networkStressMutex.Lock()
	latency := activeLatencyMs
	latencyExpires := latencyExpiry
	loss := activePacketLoss
	lossExpires := packetLossExpiry
	networkStressMutex.Unlock()

	now := time.Now()
	if now.Before(latencyExpires) && latency > 0 {
		// Delay the request processing.
		time.Sleep(time.Duration(latency) * time.Millisecond)
	}
	if now.Before(lossExpires) && loss > 0 {
		// Simulate packet loss: drop the request with the given probability.
		if rand.Intn(100) < loss {
			c.AbortWithStatusJSON(503, gin.H{
				"error":        "SERVICE_UNAVAILABLE",
				"message":      "simulated packet loss, request dropped",
				"requested_at": time.Now().UTC().Format(time.RFC3339Nano),
			})
			return
		}
	}
	c.Next()
}
