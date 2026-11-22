package util

import (
	"log"
	"os"
	"path/filepath"
	"strings"

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
	logLevel := configuredLogLevel()
	settings := loadLogSettings()

	// Set up console output. Container collectors read stdout, so the default is
	// the same structured JSON the file receives.
	consoleOutput := zapcore.Lock(os.Stdout)
	fileOutput, err := openLogFileOutput(settings)
	if err != nil {
		log.Fatal("failed to create log file: ", err)
	}

	consoleEncoder := consoleLogEncoder(settings.ConsoleFormat)
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

func configuredLogLevel() zapcore.Level {
	logLevelStr := firstNonEmpty(GetConfigByKey("logger.level"), os.Getenv("LOG_LEVEL"), "info")

	var logLevel zapcore.Level
	if err := logLevel.Set(logLevelStr); err != nil {
		// Fallback to info level if invalid
		log.Printf("Invalid log level '%s', falling back to info", logLevelStr)
		return zapcore.InfoLevel
	}
	return logLevel
}

func consoleLogEncoder(format string) zapcore.Encoder {
	if format == "text" {
		return zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}
	return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
}

// openLogFileOutput builds the rolling file sink, falling back to the default
// directory when the configured one cannot be created.
func openLogFileOutput(settings logSettings) (zapcore.WriteSyncer, error) {
	writer, err := newRollingWriter(settings)
	if err != nil && settings.Dir != defaultLogDir {
		log.Printf("Failed to create configured log directory '%s': %v, falling back to '%s'", settings.Dir, err, defaultLogDir)
		settings.Dir = defaultLogDir
		writer, err = newRollingWriter(settings)
	}
	if err != nil {
		return nil, err
	}
	return zapcore.Lock(writer), nil
}

func configuredLogFolder() string {
	if value := GetConfigByKey("logger.file"); value != "" {
		return trimLogFileToDirectory(value)
	}
	if value := os.Getenv("LOG_DIR"); value != "" {
		return trimLogFileToDirectory(value)
	}
	return trimLogFileToDirectory(os.Getenv("LOG_FOLDER"))
}

// trimLogFileToDirectory keeps the older LOG_FOLDER=<dir>/<name>.log spelling
// working now that the directory and the file-name template are separate keys.
func trimLogFileToDirectory(value string) string {
	if value == "" || !strings.HasSuffix(strings.ToLower(value), ".log") {
		return value
	}
	if directory := filepath.Dir(value); directory != "." {
		return directory
	}
	return defaultLogDir
}
