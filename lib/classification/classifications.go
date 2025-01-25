package classification

import (
	"bufio"
	"encoding/csv"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/3ideas/psasim/lib/compdb"
)

type Classification struct {
	Type        string
	ID          compdb.ComponentClassIndex
	Description string
}

type Classifications struct {
	Classifications map[string]map[compdb.ComponentClassIndex]Classification
}

func (cl *Classifications) GetClasses(typeStr string) (map[compdb.ComponentClassIndex]Classification, bool) {
	classes, ok := cl.Classifications[typeStr]
	return classes, ok
}

// ReadClassifications reads the classification file and returns a map of classifications
// each line should consist of 3 fields: Type, ID, Description
// anything after // is ignored as comments
// blank lines or lines starting with // are ignored
func ReadClassifications(fileName string) (*Classifications, error) {

	classifications := &Classifications{Classifications: make(map[string]map[compdb.ComponentClassIndex]Classification)}
	if fileName == "" {
		return classifications, nil
	}

	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		if lineNumber == 1 {
			continue // skip the first line
		}
		line := scanner.Text()
		commentIndex := strings.Index(line, "//")
		if commentIndex != -1 {
			line = strings.TrimSpace(line[:commentIndex])
		}
		if len(line) == 0 {
			continue
		}
		reader := csv.NewReader(strings.NewReader(line))
		record, err := reader.Read()
		if err != nil {
			slog.Warn("Error reading classification record, skipping ", "line", line, "error", err)
			continue
		}

		if len(record) == 0 {
			continue
		}

		if len(record) != 3 {
			slog.Warn("Invalid classification record", "line", record)
		}

		// if record[0] == "Type" || record[0] == "type" || record[0] == "TYPE" {
		// 	continue
		// }

		classificationID, err := strconv.Atoi(record[1])
		if err != nil {
			slog.Error("Invalid classification ID", "file", fileName, "line", record, "error", err)
			continue
		}
		classification := Classification{Type: record[0], ID: compdb.ComponentClassIndex(classificationID), Description: record[2]}
		if _, ok := classifications.Classifications[classification.Type]; !ok {
			classifications.Classifications[classification.Type] = make(map[compdb.ComponentClassIndex]Classification)
		}
		classifications.Classifications[classification.Type][classification.ID] = classification
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return classifications, nil
}
