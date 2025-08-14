package utils

import (
	"log"
	"os"
	"github.com/GoCOMA/Favus/internal/config"
)

var LogFilePath = config.LogFilePath

var logFile *os.File
var logger *log.Logger

func init() {
	var err error
	logFile, err = os.OpenFile(LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger = log.New(os.Stdout, "[FAVUS] ", log.LstdFlags)
		return
	}
	logger = log.New(logFile, "[FAVUS] ", log.LstdFlags)
}

func LogInfo(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[INFO] "+format, args...)
	}
	// fmt.Printf(format+"\n", args...)
}

func LogError(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] "+format, args...)
	}
	// fmt.Printf(format+"\n", args...)
}

func LogWarning(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[WARNING] "+format, args...)
	}
	// fmt.Printf(format+"\n", args...)
}

// Error is an alias for LogError for backward compatibility
func Error(format string, args ...interface{}) {
	LogError(format, args...)
}

// Info is an alias for LogInfo for backward compatibility
func Info(format string, args ...interface{}) {
	LogInfo(format, args...)
}

func closeLogFile() {
	if logFile != nil {
		logFile.Close()
	}
}
