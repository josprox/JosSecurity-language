package core

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	// GlobalLogger instance
	GlobalLogger *Logger
	loggerOnce   sync.Once
)

// Logger handles logging to file and stdout
type Logger struct {
	file *os.File
	mu   sync.Mutex
}

// InitLogger initializes the global logger
func InitLogger() {
	loggerOnce.Do(func() {
		f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening log.txt: %v\n", err)
			return
		}
		GlobalLogger = &Logger{
			file: f,
		}
	})
}

// LogError writes an error message to log.txt and stdout
func LogError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMsg := fmt.Sprintf("[%s] [ERROR] %s\n", timestamp, msg)

	// Write to Stdout
	fmt.Print(logMsg)

	// Write to File
	if GlobalLogger != nil && GlobalLogger.file != nil {
		GlobalLogger.mu.Lock()
		defer GlobalLogger.mu.Unlock()
		GlobalLogger.file.WriteString(logMsg)
	}
}

// LogInfo writes an info message to log.txt and stdout
func LogInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMsg := fmt.Sprintf("[%s] [INFO] %s\n", timestamp, msg)

	// Write to Stdout
	fmt.Print(logMsg)

	// Write to File
	if GlobalLogger != nil && GlobalLogger.file != nil {
		GlobalLogger.mu.Lock()
		defer GlobalLogger.mu.Unlock()
		GlobalLogger.file.WriteString(logMsg)
	}
}

// CloseLogger closes the log file
func CloseLogger() {
	if GlobalLogger != nil && GlobalLogger.file != nil {
		GlobalLogger.file.Close()
	}
}
