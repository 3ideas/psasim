package compare

import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/3ideas/psasim/lib/alarmstatetext"
)

type PointIDTextState struct {
	PointID string
	alarmstatetext.TextState
}

type StateTextTableByPointType struct {
	TextStateTable map[string]*PointIDTextState
}

type StateFixup struct {
	stateTextTable map[string]*StateTextTableByPointType
}

func NewStateFixup() *StateFixup {
	return &StateFixup{stateTextTable: make(map[string]*StateTextTableByPointType)}
}

func (sf *StateFixup) AddStateTextTable(stateTextTable alarmstatetext.TextStateTable, pointType string) {
	sf.stateTextTable[stateTextTable.StateTextTable] = &StateTextTableByPointType{TextStateTable: make(map[string]*PointIDTextState)}
}

func (sf *StateFixup) AddPointIDTextState(stateTextTable string, stateTextIndex int, pointID string, eToken5 string, alias string) error {

	// check if stateTextTable exists
	if _, ok := sf.stateTextTable[stateTextTable]; !ok {
		sf.stateTextTable[stateTextTable] = &StateTextTableByPointType{TextStateTable: make(map[string]*PointIDTextState)}
	}

	// check if pointID exists
	if _, ok := sf.stateTextTable[stateTextTable].TextStateTable[pointID]; !ok {
		textState, err := sf.NewTextState(stateTextTable, pointID)
		if err != nil {
			return err
		}

		sf.stateTextTable[stateTextTable].TextStateTable[pointID] = &PointIDTextState{PointID: pointID, TextState: *textState}
	}

	pointIDTextState := sf.stateTextTable[stateTextTable].TextStateTable[pointID]

	if stateTextIndex >= len(pointIDTextState.TextState.StateText) || stateTextIndex < 0 {
		return fmt.Errorf("state text index out of range: %d, %s", stateTextIndex, pointID)
	}

	if pointIDTextState.TextState.StateText[stateTextIndex] == "" {
		fmt.Printf("Adding state text for alias: %s, Table: %s, PointID: %s, eToken5: %s, StateTextIndex: %d\n", alias, stateTextTable, pointID, eToken5, stateTextIndex)
		slog.Info("Adding state text", "alias", alias, "table", stateTextTable, "pointID", pointID, "eToken5", eToken5, "stateTextIndex", stateTextIndex)
		pointIDTextState.TextState.StateText[stateTextIndex] = eToken5
	} else {
		if pointIDTextState.TextState.StateText[stateTextIndex] != eToken5 {
			slog.Error("State text mismatch", "alias", alias, "table", stateTextTable, "pointID", pointID, "eToken5", eToken5, "stateTextIndex", stateTextIndex, "current", pointIDTextState.TextState.StateText[stateTextIndex], "want", eToken5)
			return fmt.Errorf("state text mismatch for alias: %s, Table: %s, PointID: %s, eToken5: %s, StateTextIndex: %d Current: %s,  want %s\n", alias, stateTextTable, pointID, eToken5, stateTextIndex, pointIDTextState.TextState.StateText[stateTextIndex], eToken5)
		}
	}
	return nil
}

func (sf *StateFixup) NewTextState(stateTextTable string, pointID string) (*alarmstatetext.TextState, error) {

	stateLen := 0
	if stateTextTable == "SD State Text Table" {
		stateLen = 4
	} else if stateTextTable == "DD State Text Table" {
		stateLen = 8
	} else if stateTextTable == "SD Alarm State Text Table" {
		stateLen = 2
	} else if stateTextTable == "DD Alarm State Text Table" {
		stateLen = 4
	} else {
		return &alarmstatetext.TextState{}, fmt.Errorf("invalid state text table")
	}

	textStateArray := make([]string, stateLen)

	return &alarmstatetext.TextState{TableName: stateTextTable, Index: 0, StateText: textStateArray, Description: pointID}, nil
}

func (sf *StateFixup) DisplayAll() {
	stateTextTables := make([]string, 0, len(sf.stateTextTable))
	for k := range sf.stateTextTable {
		stateTextTables = append(stateTextTables, k)
	}
	sort.Strings(stateTextTables) // Sort the state text table keys

	for _, stateTextTable := range stateTextTables {
		fmt.Println(stateTextTable)
		pointIDs := make([]string, 0, len(sf.stateTextTable[stateTextTable].TextStateTable))
		for k := range sf.stateTextTable[stateTextTable].TextStateTable {
			pointIDs = append(pointIDs, k)
		}
		sort.Strings(pointIDs) // Sort the point ID keys

		for _, pointID := range pointIDs {
			pointIDTextState := sf.stateTextTable[stateTextTable].TextStateTable[pointID]
			fmt.Println(pointIDTextState.TextState)
		}
	}
}

func (sf *StateFixup) FixupAll() {
	for stateTextTable, stateTextTableByPointType := range sf.stateTextTable {
		fmt.Println(stateTextTable)
		for pointID, pointIDTextState := range stateTextTableByPointType.TextStateTable {
			fmt.Println(pointID, pointIDTextState.TextState)
		}
	}
}
