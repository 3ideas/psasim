package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/3ideas/psasim/lib/compare"
	"github.com/3ideas/psasim/lib/compdb"
	"github.com/3ideas/psasim/lib/loglevel"
	"github.com/3ideas/psasim/lib/namer_service/namer_client"
	"github.com/3ideas/psasim/lib/namer_service/namer_server"
	"github.com/3ideas/psasim/lib/namerif"
	"github.com/3ideas/psasim/lib/psalerts"
)

func main() {

	logFilename := flag.String("log", "", "log file")
	logLevel := flag.String("loglevel", "info", "log level")

	psalertsFile := flag.String("psalerts", "", "PSAlerts CSV file")
	dbFile := flag.String("db", "", "database file")
	server := flag.Bool("server", false, "run as server")
	useNameService := flag.Bool("usenameservice", false, "use name service for name resolution (rather than dn)")
	checkaliases := flag.Bool("checkaliases", false, "check aliases")
	dumpNames := flag.String("dumpnames", "", "dump names to file")

	comparisonFile := flag.String("comparisonfile", "", "comparison file")

	flag.Parse()

	logFile := loglevel.SetLogger(*logFilename, *logLevel)
	if logFile != nil {
		defer logFile.Close()
	}

	var alarms *psalerts.PSAlerts
	var err error
	if *psalertsFile != "" {
		alarms, err = psalerts.ReadPSAlerts(*psalertsFile)
		if err != nil {
			log.Fatal("Error reading file:", err)
		}
	}

	var compDb *compdb.ComponentDb
	if *dbFile != "" {
		startTime := time.Now() // Start timer
		fmt.Printf("Reading namer from %s\n", *dbFile)
		slog.Info("Reading namer from ", "file", *dbFile)
		compDb, err = compdb.LoadCompDb(*dbFile)
		if err != nil {
			fmt.Printf("Error reading database: %s does file exist? %s\n", *dbFile, err)
			slog.Error("Error reading database", "Error", err, "file", *dbFile)
			log.Fatal("Error reading database", err)
		}
		duration := time.Since(startTime)                          // Calculate duration
		slog.Info("Completed reading namer", "duration", duration) // Log duration
		fmt.Printf("Completed reading namer in %s\n", duration)    // Print duration
	}

	if *dumpNames != "" {
		compDb.DumpNames(*dumpNames)
	}

	var alarmComparison *compare.AlarmsComparison
	var eterraToPO *compare.EterraToPO
	if *comparisonFile != "" {
		alarmComparison, err = compare.ReadAlarmComparison(*comparisonFile)
		if err != nil {
			log.Fatal("Error reading comparison file", err)
		}

		eterraToPO = alarmComparison.EterraToPO()
	}

	// Do we need to run the server?
	if *server && compDb != nil {
		server := namer_server.NewNameServer(compDb)
		slog.Info("name server started")
		fmt.Printf("name server started\n")
		err := server.StartServer()
		if err != nil {
			slog.Error("Error starting name server", "Error", err)
			return
		}
	}

	var nameserver *namer_client.NameClient
	if *useNameService {
		nameserver, err = namer_client.Connect()
		if err != nil {
			slog.Error("Error connecting to name server", "Error", err)
			return
		}
		noOfChanges, err := nameserver.GetNumberOfChanges()
		if err != nil {
			slog.Error("Error getting number of changes", "Error", err)
			return
		}
		slog.Info("Number of changes", "Number of changes", noOfChanges)
		fmt.Printf("Number of changes: %d\n", noOfChanges)
		if noOfChanges > 0 { // rollback any changes
			err = nameserver.RollbackAll()
			if err != nil {
				fmt.Printf("Error rolling back all: %s\n", err)
				slog.Error("Error rolling back all", "Error", err)
				return
			}
			slog.Info("Undid all")
			noOfChanges, err = nameserver.GetNumberOfChanges()
			if err != nil {
				slog.Error("Error getting number of changes", "Error", err)
				return
			}
			slog.Info("Number of changes", "Number of changes", noOfChanges)
			fmt.Printf("Number of changes after rolling back all: %d\n", noOfChanges)
		}
		defer nameserver.Close()
	}

	if *useNameService && nameserver == nil {
		fmt.Println("No name server found")
		log.Fatal("No name server found")
	}
	if nameserver == nil && compDb == nil {
		fmt.Println("No name server or database found")
		log.Fatal("No name server or database found")
	}

	var nameChecker namerif.NameService
	if nameserver != nil {
		nameChecker = nameserver
	} else {
		nameChecker = compDb
	}

	if *checkaliases {
		psalerts.CheckAliases(nameChecker)
	}

	if alarms == nil {
		fmt.Println("No alarms to process")
		log.Fatal("No alarms to process")
	}

	alarms.PrintAlarmCounts()
	alarms.ResolveAliases(nameChecker, eterraToPO)

}
