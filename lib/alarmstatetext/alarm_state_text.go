package alarmstatetext

import (
	"encoding/csv"
	"os"
)

type AlarmStateText struct {
	PointID    string
	StateTable string
	Index      string
	Comments   string
}

func ReadAlarmStateText(filePath string) (map[string]AlarmStateText, error) {
	alarmStateTextMappings := make(map[string]AlarmStateText)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Read all rows from the CSV
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Skip the header row and read the data
	for _, record := range records[1:] { // Start from index 1 to skip header
		if len(record) < 4 {
			continue // Skip rows that don't have enough columns
		}
		mapping := AlarmStateText{
			PointID:    record[0],
			StateTable: record[1],
			Index:      record[2],
			Comments:   record[3],
		}
		alarmStateTextMappings[mapping.PointID] = mapping
	}

	return alarmStateTextMappings, nil
}
