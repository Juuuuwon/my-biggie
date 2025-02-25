package main

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FileWritePayload defines the JSON payload for heavy file write stress.
type FileWritePayload struct {
	FileSize       DuckInt `json:"file_size"`       // Size in bytes per file.
	FileCount      DuckInt `json:"file_count"`      // Number of files per interval.
	MaintainSecond DuckInt `json:"maintain_second"` // Total duration.
	Async          bool    `json:"async"`           // Run in background if true.
	IntervalSecond DuckInt `json:"interval_second"` // Interval between writes.
}

// FileWriteHandler handles POST /stress/filesystem/write.
func FileWriteHandler(c *gin.Context) {
	var payload FileWritePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}

	fileSize := int(payload.FileSize)
	fileCount := int(payload.FileCount)
	maintainSec := int(payload.MaintainSecond)
	intervalSec := int(payload.IntervalSecond)

	if payload.Async {
		go runFileWriteStress(fileSize, fileCount, maintainSec, intervalSec)
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "file write stress started",
			"file_size":       fileSize,
			"file_count":      fileCount,
			"maintain_second": maintainSec,
			"interval_second": intervalSec,
		})
	} else {
		runFileWriteStress(fileSize, fileCount, maintainSec, intervalSec)
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "file write stress completed",
			"file_size":       fileSize,
			"file_count":      fileCount,
			"maintain_second": maintainSec,
			"interval_second": intervalSec,
		})
	}
}

func runFileWriteStress(fileSize, fileCount, maintainSec, intervalSec int) {
	// Determine temporary directory.
	tmpDir := os.TempDir()
	endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
	interval := time.Duration(intervalSec) * time.Second

	for time.Now().Before(endTime) {
		for i := 0; i < fileCount; i++ {
			// Create a temporary file name.
			filename := filepath.Join(tmpDir, "biggie_write_"+strconv.FormatInt(time.Now().UnixNano(), 10)+"_"+strconv.Itoa(i)+".tmp")
			data := make([]byte, fileSize)
			// Fill data with random bytes.
			rand.Read(data)
			// Write data to file.
			err := ioutil.WriteFile(filename, data, 0644)
			if err != nil {
				log("failed to write file", zap.String("file", filename), zap.Error(err))
			} else {
				// Optionally remove file immediately to avoid disk fill.
				os.Remove(filename)
			}
		}
		time.Sleep(interval)
	}
	log("File write stress completed", zap.Int("file_size", fileSize), zap.Int("file_count", fileCount))
}

// FileReadPayload defines the JSON payload for heavy file read stress.
type FileReadPayload struct {
	FilePath       string  `json:"file_path"`       // File to read.
	MaintainSecond DuckInt `json:"maintain_second"` // Duration.
	Async          bool    `json:"async"`           // Background if true.
	ReadFrequency  DuckInt `json:"read_frequency"`  // Reads per interval.
	IntervalSecond DuckInt `json:"interval_second"` // Interval duration.
}

// FileReadHandler handles POST /stress/filesystem/read.
func FileReadHandler(c *gin.Context) {
	var payload FileReadPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorJSON(c, http.StatusBadRequest, "INVALID_PAYLOAD", err.Error())
		return
	}

	maintainSec := int(payload.MaintainSecond)
	readFreq := int(payload.ReadFrequency)
	intervalSec := int(payload.IntervalSecond)
	filePath := payload.FilePath

	if payload.Async {
		go runFileReadStress(filePath, maintainSec, readFreq, intervalSec)
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "file read stress started",
			"file_path":       filePath,
			"maintain_second": maintainSec,
			"read_frequency":  readFreq,
			"interval_second": intervalSec,
		})
	} else {
		runFileReadStress(filePath, maintainSec, readFreq, intervalSec)
		ResponseJSON(c, http.StatusOK, gin.H{
			"message":         "file read stress completed",
			"file_path":       filePath,
			"maintain_second": maintainSec,
			"read_frequency":  readFreq,
			"interval_second": intervalSec,
		})
	}
}

func runFileReadStress(filePath string, maintainSec, readFreq, intervalSec int) {
	endTime := time.Now().Add(time.Duration(maintainSec) * time.Second)
	interval := time.Duration(intervalSec) * time.Second

	for time.Now().Before(endTime) {
		for i := 0; i < readFreq; i++ {
			_, err := ioutil.ReadFile(filePath)
			if err != nil {
				log("failed to read file", zap.String("file", filePath), zap.Error(err))
			}
		}
		time.Sleep(interval)
	}
	log("File read stress completed", zap.String("file_path", filePath))
}
