package main

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	lorem "github.com/drhodes/golorem"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaHeavyPayload defines the payload for the heavy Kafka produce using a single producer.
type KafkaHeavyPayload struct {
	Messages           string  `json:"messages"` // If empty, a lorem ipsum message is generated automatically.
	MaintainSecond     DuckInt `json:"maintain_second"`
	Async              bool    `json:"async"`
	ProducePerInterval DuckInt `json:"produce_per_interval"`
	IntervalSecond     DuckInt `json:"interval_second"`
}

// KafkaMultiHeavyPayload defines the payload for heavy Kafka produce using multiple producers.
type KafkaMultiHeavyPayload struct {
	Messages           string  `json:"messages"` // If empty, a lorem ipsum message is generated automatically.
	MaintainSecond     DuckInt `json:"maintain_second"`
	Async              bool    `json:"async"`
	ConnectionCounts   DuckInt `json:"connection_counts"`
	ProducePerInterval DuckInt `json:"produce_per_interval"`
	IntervalSecond     DuckInt `json:"interval_second"`
}

// KafkaConnectionPayload defines the payload for simulating heavy Kafka connections.
type KafkaConnectionPayload struct {
	MaintainSecond      DuckInt `json:"maintain_second"`
	Async               bool    `json:"async"`
	ConnectionCounts    DuckInt `json:"connection_counts"`
	IncreasePerInterval DuckInt `json:"increase_per_interval"`
	IntervalSecond      DuckInt `json:"interval_second"`
}

// getKafkaWriter creates and returns a new kafka-go Writer using configuration from GetKafkaConfig.
func getKafkaWriter() (*kafka.Writer, error) {
	cfg, err := GetKafkaConfig()
	if err != nil {
		return nil, err
	}
	// cfg.Servers is already a []string, so use it directly.
	writerConfig := kafka.WriterConfig{
		Brokers:  cfg.Servers,
		Topic:    cfg.Topic,
		Balancer: &kafka.LeastBytes{},
		// Set a default Dialer; this can be overridden below if TLS is enabled.
		Dialer: &kafka.Dialer{},
	}
	if cfg.TLSEnabled {
		writerConfig.Dialer = &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			TLS: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	return kafka.NewWriter(writerConfig), nil
}

// generateLoremIpsum uses the golorem library to generate a lorem ipsum text.
// It generates a text with a random number of words between 10 and 20.
func generateLoremIpsum() string {
	// Generate a lorem ipsum text with 10 to 20 words.
	return lorem.Word(10, 20)
}

// KafkaHeavyHandler handles POST /kafka/heavy.
// It uses a single producer to send messages at a controlled rate for maintain_second seconds.
func KafkaHeavyHandler(c *gin.Context) {
	var payload KafkaHeavyPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	producePerInterval := int(payload.ProducePerInterval)
	intervalSec := int(payload.IntervalSecond)
	// Use provided message or auto-generate using lorem ipsum if empty.
	messageContent := payload.Messages
	if messageContent == "" {
		messageContent = generateLoremIpsum()
	}

	writer, err := getKafkaWriter()
	if err != nil {
		ErrorJSON(c, 500, "KAFKA_ERROR", err.Error())
		return
	}

	stressFunc := func() {
		endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
		for time.Now().Before(endTime) {
			messages := make([]kafka.Message, 0, producePerInterval)
			for i := 0; i < producePerInterval; i++ {
				messages = append(messages, kafka.Message{
					Key:   []byte(fmt.Sprintf("key-%d", i)),
					Value: []byte(messageContent),
				})
			}
			if err := writer.WriteMessages(c, messages...); err != nil {
				fmt.Println("Kafka heavy produce failed", zap.Error(err))
			}
			time.Sleep(time.Duration(intervalSec) * time.Second)
		}
		writer.Close()
		fmt.Println("Kafka heavy produce (single producer) completed", zap.Int("duration_sec", maintainSec))
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":              "Kafka heavy produce started",
			"maintain_second":      maintainSec,
			"produce_per_interval": producePerInterval,
			"interval_second":      intervalSec,
			"messages":             messageContent,
		})
	} else {
		stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":              "Kafka heavy produce completed",
			"maintain_second":      maintainSec,
			"produce_per_interval": producePerInterval,
			"interval_second":      intervalSec,
			"messages":             messageContent,
		})
	}
}

// KafkaMultiHeavyHandler handles POST /kafka/multi_heavy.
// It spawns multiple producer connections (as specified by connection_counts)
// with each producer sending messages at the given rate concurrently.
func KafkaMultiHeavyHandler(c *gin.Context) {
	var payload KafkaMultiHeavyPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	producePerInterval := int(payload.ProducePerInterval)
	intervalSec := int(payload.IntervalSecond)
	connectionCounts := int(payload.ConnectionCounts)
	// Use provided message or auto-generate using lorem ipsum if empty.
	messageContent := payload.Messages
	if messageContent == "" {
		messageContent = generateLoremIpsum()
	}

	stressFunc := func() {
		var wg sync.WaitGroup
		for i := 0; i < connectionCounts; i++ {
			wg.Add(1)
			go func(connNum int) {
				defer wg.Done()
				writer, err := getKafkaWriter()
				if err != nil {
					fmt.Println("Kafka multi heavy writer creation failed", zap.Int("conn", connNum), zap.Error(err))
					return
				}
				endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
				for time.Now().Before(endTime) {
					messages := make([]kafka.Message, 0, producePerInterval)
					for j := 0; j < producePerInterval; j++ {
						messages = append(messages, kafka.Message{
							Key:   []byte(fmt.Sprintf("conn-%d-key-%d", connNum, j)),
							Value: []byte(messageContent),
						})
					}
					if err := writer.WriteMessages(c, messages...); err != nil {
						fmt.Println("Kafka multi heavy produce failed", zap.Int("conn", connNum), zap.Error(err))
					}
					time.Sleep(time.Duration(intervalSec) * time.Second)
				}
				writer.Close()
			}(i)
		}
		wg.Wait()
		fmt.Println("Kafka multi heavy produce completed", zap.Int("producers", connectionCounts))
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":              "Kafka multi heavy produce started",
			"maintain_second":      maintainSec,
			"produce_per_interval": producePerInterval,
			"interval_second":      intervalSec,
			"connection_counts":    connectionCounts,
			"messages":             messageContent,
		})
	} else {
		stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":              "Kafka multi heavy produce completed",
			"maintain_second":      maintainSec,
			"produce_per_interval": producePerInterval,
			"interval_second":      intervalSec,
			"connection_counts":    connectionCounts,
			"messages":             messageContent,
		})
	}
}

// KafkaConnectionHandler handles POST /kafka/connection.
// It gradually establishes multiple producer connections until reaching the target count,
// maintains them open for the specified duration, and then closes them.
func KafkaConnectionHandler(c *gin.Context) {
	var payload KafkaConnectionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, 400, "INVALID_PAYLOAD", err.Error())
		return
	}
	maintainSec := int(payload.MaintainSecond)
	connectionCounts := int(payload.ConnectionCounts)
	increasePerInterval := int(payload.IncreasePerInterval)
	intervalSec := int(payload.IntervalSecond)

	stressFunc := func() {
		var writers []*kafka.Writer
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
					writer, err := getKafkaWriter()
					if err != nil {
						fmt.Println("Kafka connection stress writer creation failed", zap.Error(err))
						continue
					}
					mu.Lock()
					writers = append(writers, writer)
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
		for _, writer := range writers {
			writer.Close()
		}
		mu.Unlock()
		fmt.Println("Kafka connection stress completed", zap.Int("producers", currentCount))
	}

	if payload.Async {
		go stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":               "Kafka connection stress started",
			"maintain_second":       maintainSec,
			"connection_counts":     connectionCounts,
			"increase_per_interval": increasePerInterval,
			"interval_second":       intervalSec,
		})
	} else {
		stressFunc()
		ResponseJSON(c, 200, gin.H{
			"message":               "Kafka connection stress completed",
			"maintain_second":       maintainSec,
			"connection_counts":     connectionCounts,
			"increase_per_interval": increasePerInterval,
			"interval_second":       intervalSec,
		})
	}
}
