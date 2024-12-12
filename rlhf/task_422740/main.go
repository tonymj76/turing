package main

import (
	"fmt"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

func main() {
	// Set up logger
	logger := logrus.New()

	// Configure lumberjack for log rotation
	logger.SetOutput(&lumberjack.Logger{
		Filename:   "application.log", // Log file path
		MaxSize:    10,                // Max size in MB before rotation
		MaxBackups: 5,                 // Number of backups to keep
		MaxAge:     30,                // Max age in days before deletion
		Compress:   true,              // Compress rotated files
	})

	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Simulate log messages being generated over time
	for i := 0; i < 100; i++ {
		logger.Info(fmt.Sprintf("Log message %d", i))
		time.Sleep(1 * time.Second)
	}
}
