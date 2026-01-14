package util

import (
	"fmt"
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	consoleLogger        *zap.Logger
	sugaredConsoleLogger *zap.SugaredLogger
	Logger               = getLogger()
)

// Initialize the logger instances if they don't exist
func initLogger() {
	if consoleLogger == nil {
		consoleLogger = initConsoleLogger()
	}
	if sugaredConsoleLogger == nil {
		sugaredConsoleLogger = consoleLogger.Sugar()
	}

	consoleLogger.Debug("loggers initialized successfully")
}

// GetLogger returns the initialized sugared logger instance
// If the logger is not initialized, it will be initialized first
func getLogger() *zap.SugaredLogger {
	if sugaredConsoleLogger == nil {
		initLogger()
	}
	return sugaredConsoleLogger
}

func initConsoleLogger() *zap.Logger {
	// Get log level from config, default to info
	logLevelStr := GetConfigByKey("logger.level")
	if logLevelStr == "" {
		logLevelStr = "info"
	}

	// Convert string to zapcore.Level
	var logLevel zapcore.Level
	err := logLevel.Set(logLevelStr)
	if err != nil {
		// Fallback to info level if invalid
		log.Printf("Invalid log level '%s', falling back to info", logLevelStr)
		logLevel = zapcore.InfoLevel
	}

	// Set up console output
	consoleOutput := zapcore.Lock(os.Stdout)
	// Set up file output
	fileOutput := zapcore.Lock(zapcore.AddSync(createLogFile()))

	// Set up log encoders
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	// Create cores for different outputs
	consoleCore := zapcore.NewCore(consoleEncoder, consoleOutput, logLevel)
	fileCore := zapcore.NewCore(fileEncoder, fileOutput, logLevel)

	// Combine console and file outputs
	logger := zap.New(zapcore.NewTee(consoleCore, fileCore), zap.AddCaller())

	// Note: Logger sync should be called at program exit
	// defer consoleLogger.Sync()

	return logger
}

func createLogFile() *os.File {
	// Get the log folder from config, default to ./logs
	logFolder := GetConfigByKey("logger.file")

	// Extract folder path if a full filename was provided
	if logFolder != "" {
		// If the path has a file extension, extract just the directory
		if len(logFolder) > 4 && logFolder[len(logFolder)-4:] == ".log" {
			// Check if it's just a filename without directory
			if logFolder == "./cashlenx.log" || logFolder == "cashlenx.log" {
				logFolder = "./logs"
			} else {
				// Extract directory part
				lastSlash := -1
				for i := len(logFolder) - 1; i >= 0; i-- {
					if logFolder[i] == '/' || logFolder[i] == '\\' {
						lastSlash = i
						break
					}
				}
				if lastSlash >= 0 {
					logFolder = logFolder[:lastSlash]
				} else {
					logFolder = "."
				}
			}
		}
	} else {
		logFolder = "./logs"
	}

	// Create the log folder if it doesn't exist, with fallback to default
	if err := os.MkdirAll(logFolder, 0755); err != nil {
		// Fallback to default logs folder if configured one fails
		log.Printf("Failed to create configured log folder '%s': %v, falling back to default './logs/'", logFolder, err)
		defaultFolder := "./logs"
		if err := os.MkdirAll(defaultFolder, 0755); err != nil {
			log.Fatalf("Failed to create both configured and default log folders: %v", err)
		}
		logFolder = defaultFolder
	}

	// Always use date-based filename in the configured folder
	currentDate := time.Now().Format("20060102")
	logFilePath := fmt.Sprintf("%s/cashlenx_%s.log", logFolder, currentDate)

	// Open file with append mode if it exists, create if it doesn't
	file, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("failed to create log file: ", err)
	}
	return file
}
