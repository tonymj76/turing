package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

// LoggerConfig holds the logger configuration and allows dynamic updates.
type LoggerConfig struct {
	mu      sync.RWMutex
	logger  *logrus.Logger
	logFile *lumberjack.Logger
	level   logrus.Level
}

// SetLogLevel dynamically updates the log level.
func (config *LoggerConfig) SetLogLevel(level logrus.Level) {
	config.mu.Lock()
	defer config.mu.Unlock()
	config.level = level
	config.logger.SetLevel(level)
}

// GetLogLevel retrieves the current log level.
func (config *LoggerConfig) GetLogLevel() logrus.Level {
	config.mu.RLock()
	defer config.mu.RUnlock()
	return config.level
}

func main() {
	// Initialize LoggerConfig
	logFile := &lumberjack.Logger{
		Filename:   "application.log",
		MaxSize:    10,   // Max size in MB
		MaxBackups: 5,    // Max number of backup files
		MaxAge:     30,   // Max age in days
		Compress:   true, // Compress the rotated logs
	}
	logger := logrus.New()
	logger.SetOutput(logFile)
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logger.SetLevel(logrus.InfoLevel)

	config := &LoggerConfig{
		logger:  logger,
		logFile: logFile,
		level:   logrus.InfoLevel,
	}

	// Start HTTP server for dynamic log level updates
	go startHTTPServer(config)

	// Simulate log generation
	for i := 0; i < 100; i++ {
		config.logger.Infof("Log message %d at level: %s", i, config.GetLogLevel().String())
		time.Sleep(2 * time.Second)
	}
}

func startHTTPServer(config *LoggerConfig) {
	http.HandleFunc("/loglevel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Get the current log level
			level := config.GetLogLevel()
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"log_level": level.String()})
		} else if r.Method == http.MethodPost {
			// Set a new log level
			var req map[string]string
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			levelStr, ok := req["log_level"]
			if !ok {
				http.Error(w, "Missing log_level field", http.StatusBadRequest)
				return
			}

			level, err := logrus.ParseLevel(levelStr)
			if err != nil {
				http.Error(w, "Invalid log level", http.StatusBadRequest)
				return
			}

			config.SetLogLevel(level)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "log_level": level.String()})
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("HTTP server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("HTTP server error: %v\n", err)
	}
}
