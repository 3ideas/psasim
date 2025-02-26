package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"

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
	resolvedAlarmsFile := flag.String("resolvedalarmsfile", "", "resolved alarms file")
	unresolvedAlarmsFile := flag.String("unresolvedalarmsfile", "", "unresolved alarms file")

	flag.Parse()

	logFile := loglevel.SetLogger(*logFilename, *logLevel)
	if logFile != nil {
		defer logFile.Close()
	}

	var alerts *psalerts.PSAlerts
	var err error
	if *psalertsFile != "" {
		alerts, err = psalerts.ReadPSAlerts(*psalertsFile)
		if err != nil {
			fmt.Printf("Error reading file: %s\n", err)
			slog.Error("Error reading file", "Error", err)
			log.Fatal("Error reading file:", err)
		}
	}

	var compDb *compdb.ComponentDb
	if *dbFile != "" {
		compDb, err = compdb.ReadDB(*dbFile)
		if err != nil {
			log.Fatal("Error reading database:", err)
		}
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

	if alerts == nil {
		fmt.Println("No alarms to process")
		log.Fatal("No alarms to process")
	}

	alerts.PrintAlarmCounts()

	if *resolvedAlarmsFile != "" {
		err = alerts.ResolveAliases(nameChecker, eterraToPO, *resolvedAlarmsFile, *unresolvedAlarmsFile)
		if err != nil {
			fmt.Printf("Error resolving aliases: %s\n", err)
			slog.Error("Error resolving aliases", "Error", err)
			return
		}
	}
}
