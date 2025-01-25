package loglevel

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

var LogLevel = slog.Level(slog.LevelWarn)

// For  slog to set the level dynamically
var programLevel *slog.LevelVar

func StrToLogLevel(level string) int {

	loglevel, err := strconv.Atoi(level)
	if err != nil {
		log.Fatal(err)
	}

	return loglevel
}

const LevelError = slog.LevelError
const LevelWarn = slog.LevelWarn
const LevelInfo = slog.LevelInfo
const LevelDebug = slog.LevelDebug
const LevelTrace = slog.Level(-8)

func StrToSlogLevel(level string) slog.Level {

	logLevel := strings.ToLower(level)

	switch logLevel {
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	case "info":
		return slog.LevelInfo
	case "debug":
		return slog.LevelDebug
	case "trace":
		return LevelTrace
	}
	return slog.LevelWarn
}

func SetLogLevelStr(level string) {

	LogLevel = StrToSlogLevel(level)

	//log.SetLevel(loglevel)

	programLevel.Set(LogLevel)
}

// func SetLogLevel(level int) {

// 	LogLevel = slog.Level(level)

// 	if level == 0 {
// 		log.SetLevel(log.ErrorLevel)
// 	} else if level == 1 {
// 		log.SetLevel(log.WarnLevel)
// 	} else if level == 2 {
// 		log.SetLevel(log.InfoLevel)
// 	} else if level == 3 {
// 		log.SetLevel(log.DebugLevel)
// 	} else if level == 4 {
// 		log.SetLevel(log.TraceLevel)
// 	}
// }

func GetLogLevel() int {
	return int(LogLevel)
}

func SetLogger(filename string, logLevel string) *os.File {

	var logFile *os.File
	var err error

	if filename != "" {
		logFile, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		log.SetOutput(logFile)
	} else {
		// force colours on in vs code!
		log.SetFormatter(&log.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
	}

	programLevel = new(slog.LevelVar) // Info by default
	var logger *slog.Logger

	handler := &slog.HandlerOptions{AddSource: false, // if we want file names
		Level: programLevel}

	if logFile == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, handler))
	} else {
		//logger = slog.New(slog.NewJSONHandler(logFile, handler))

		logger = slog.New(slog.NewTextHandler(logFile, handler))
	}
	slog.SetDefault(logger)

	SetLogLevelStr(logLevel)

	// slog.Info("Debug message: %s  hi", "hello")

	slog.LogAttrs(context.Background(), LevelInfo, "Logging set",
		slog.String("filename", filename),
		slog.String("logLevel", logLevel))

	return logFile
}
