package compdb

import (
	"fmt"
	"log"
	"log/slog"
	"time"
)

func ReadDB(filename string) (*ComponentDb, error) {

	startTime := time.Now() // Start timer
	fmt.Printf("Reading namer from %s\n", filename)
	slog.Info("Reading namer from ", "file", filename)
	compDb, err := LoadCompDb(filename)
	if err != nil {
		fmt.Printf("Error reading database: %s does file exist? %s\n", filename, err)
		slog.Error("Error reading database", "Error", err, "file", filename)
		log.Fatal("Error reading database", err)
	}
	duration := time.Since(startTime)                          // Calculate duration
	slog.Info("Completed reading namer", "duration", duration) // Log duration
	fmt.Printf("Completed reading namer in %s\n", duration)    // Print duration

	return compDb, nil
}
