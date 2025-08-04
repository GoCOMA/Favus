package favus

import (
	"fmt"
	"log"
	"os"
)

// 로그 파일 경로
const LogFilePath = "./favus.log"

var logFile *os.File
var logger *log.Logger

func init() {
	var err error
	logFile, err = os.OpenFile(LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// 로그 파일 생성 실패 시 표준 출력으로 fallback
		logger = log.New(os.Stdout, "[FAVUS] ", log.LstdFlags)
		return
	}
	logger = log.New(logFile, "[FAVUS] ", log.LstdFlags)
}

func LogInfo(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[INFO] "+format, args...)
	}
	fmt.Printf(format+"\n", args...)
}

func LogError(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] "+format, args...)
	}
	fmt.Printf(format+"\n", args...)
}

func LogWarning(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[WARNING] "+format, args...)
	}
	fmt.Printf(format+"\n", args...)
}

func CloseLogFile() {
	if logFile != nil {
		logFile.Close()
	}
}
