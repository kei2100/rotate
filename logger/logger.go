package logger

import (
	"log"
	"os"
	"sync"
)

var (
	logger   Logger = log.New(os.Stderr, "", log.LstdFlags)
	loggerMu sync.RWMutex
)

// Logger interface
type Logger interface {
	// Print prints v
	Print(v ...interface{})
	// Println prints v
	Println(v ...interface{})
	// Printf prints v specified format
	Printf(format string, v ...interface{})
	// Fatal fatal prints v
	Fatal(v ...interface{})
	// Fatalln fatal prints v
	Fatalln(v ...interface{})
	// Fatalf fatal prints v specified format
	Fatalf(format string, v ...interface{})
}

// Set the logger
func Set(l Logger) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	logger = l
}

// Print calls logger.Print
func Print(v ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	logger.Print(v...)
}

// Println calls logger.Println
func Println(v ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	logger.Println(v...)
}

// Printf calls logger.Printf
func Printf(format string, v ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	logger.Printf(format, v...)
}

// Fatal calls logger.Fatal
func Fatal(v ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	logger.Fatal(v...)
}

// Fatalln calls logger.Fatalln
func Fatalln(v ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	logger.Fatalln(v...)
}

// Fatalf calls logger.Fatalf
func Fatalf(format string, v ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	logger.Fatalf(format, v...)
}
