package alarmstatetext

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/3ideas/psasim/lib/namerif"
)

type ScadaStateText struct {
	StateTextTable      string `csv:"State Text Table"`
	StateTextTableIndex string `csv:"State Text Table Index (State Index)"`
	StateAlarm          string `csv:"State Alarm"`
	state0Text          string
	state1Text          string
	state2Text          string
	state3Text          string
	state4Text          string
	state5Text          string
	state6Text          string
	state7Text          string
}

type ScadaAttributes struct {
	PointID   string `csv:"Point ID"`
	TestValue string `csv:"Test Value"`
	PToken5   string `csv:"pToken5"`
	EToken5   string `csv:"eToken5"`

	ScadaStateText
	StateTextIndex             int    `csv:"State Text Index"` // index of the text in the StateText array
	StateTextTableUpdatedIndex string `csv:"State Text Table Index (New)"`
	Error                      string
	FixupError                 string
}

func (scada *ScadaStateText) GetStateText(alias string, nameserver namerif.NameService) {

	attr, err := nameserver.GetAttributeValue(alias, "State Alarm")
	if err == nil {
		scada.StateAlarm = attr.Definition
	}

	stateTextTable := ""

	attr, err = nameserver.GetAttributeValue(alias, "State Index")
	if err == nil {
		scada.StateTextTableIndex = attr.Value
	}
	attr, err = nameserver.GetAttributeValue(alias, "State 0 Text")
	if err == nil {
		scada.state0Text = attr.Definition
	} else {
		attr, err = nameserver.GetAttributeValue(alias, "State 0 text")
		if err == nil {
			scada.state0Text = attr.Definition
		}
	}

	stateTextTable = getStateTextTable(stateTextTable, scada.state0Text)

	attr, err = nameserver.GetAttributeValue(alias, "State 1 Text")
	if err == nil {
		scada.state1Text = attr.Definition
	} else {
		attr, err = nameserver.GetAttributeValue(alias, "State 1 text")
		if err == nil {
			scada.state1Text = attr.Definition
		}
	}
	stateTextTable = getStateTextTable(stateTextTable, scada.state1Text)

	attr, err = nameserver.GetAttributeValue(alias, "State 2 Text")
	if err == nil {
		scada.state2Text = attr.Definition
	} else {
		attr, err = nameserver.GetAttributeValue(alias, "State 2 text")
		if err == nil {
			scada.state2Text = attr.Definition
		}
	}
	stateTextTable = getStateTextTable(stateTextTable, scada.state2Text)

	attr, err = nameserver.GetAttributeValue(alias, "State 3 Text")
	if err == nil {
		scada.state3Text = attr.Definition
	} else {
		attr, err = nameserver.GetAttributeValue(alias, "State 3 text")
		if err == nil {
			scada.state3Text = attr.Definition
		}
	}
	stateTextTable = getStateTextTable(stateTextTable, scada.state3Text)

	attr, err = nameserver.GetAttributeValue(alias, "State 4 Text")
	if err == nil {
		scada.state4Text = attr.Definition
	} else {
		attr, err = nameserver.GetAttributeValue(alias, "State 4 text")
		if err == nil {
			scada.state4Text = attr.Definition
		}
	}
	stateTextTable = getStateTextTable(stateTextTable, scada.state4Text)

	attr, err = nameserver.GetAttributeValue(alias, "State 5 Text")
	if err == nil {
		scada.state5Text = attr.Definition
	} else {
		attr, err = nameserver.GetAttributeValue(alias, "State 5 text")
		if err == nil {
			scada.state5Text = attr.Definition
		}
	}
	stateTextTable = getStateTextTable(stateTextTable, scada.state5Text)

	attr, err = nameserver.GetAttributeValue(alias, "State 6 Text")
	if err == nil {
		scada.state6Text = attr.Definition
	} else {
		attr, err = nameserver.GetAttributeValue(alias, "State 6 text")
		if err == nil {
			scada.state6Text = attr.Definition
		}
	}
	stateTextTable = getStateTextTable(stateTextTable, scada.state6Text)

	attr, err = nameserver.GetAttributeValue(alias, "State 7 Text")
	if err == nil {
		scada.state7Text = attr.Definition
	} else {
		attr, err = nameserver.GetAttributeValue(alias, "State 7 text")
		if err == nil {
			scada.state7Text = attr.Definition
		}
	}
	stateTextTable = getStateTextTable(stateTextTable, scada.state7Text)

	scada.StateTextTable = stateTextTable
}

func GetScadaAttributes(alias string, nameserver namerif.NameService, pointID string, testValue string, eToken5 string, pToken5 string, pointType string) (*ScadaAttributes, error) {

	scada := &ScadaAttributes{}

	scada.PointID = pointID
	scada.TestValue = testValue
	scada.PToken5 = pToken5
	scada.EToken5 = eToken5

	scada.GetStateText(alias, nameserver)

	stateTextTable := scada.StateTextTable

	// convert StateIndex to int
	stateIndex, err := strconv.Atoi(scada.StateTextTableIndex)
	if err != nil {
		return nil, fmt.Errorf("error converting StateIndex to int: %s", err)
	}
	textState, err := GetTextState(stateTextTable, stateIndex, pointType)
	if err != nil {
		scada.Error = fmt.Sprintf("%s", err)
		return scada, err
	}

	scada.StateTextIndex, err = textState.GetTextIndex(pToken5)
	if err != nil {
		scada.Error = fmt.Sprintf("%s: %s", stateTextTable, err)
		return scada, err
	}

	return scada, nil
}

// ... existing code ...
func getStateTextTable(stateTextTable, state0Text string) string {
	if stateTextTable != "" {
		return stateTextTable
	}

	// Extract the text after "State Text Tables." up to the first ]
	start := strings.Index(state0Text, "State Text Tables.")
	if start == -1 {
		return ""
	}
	start += len("State Text Tables.")
	end := strings.Index(state0Text[start:], "]")
	if end == -1 {
		return ""
	}

	return state0Text[start : start+end]
}

// Define the struct
type TextState struct {
	TableName   string
	Index       int
	StateText   []string // Array to hold State 0 to State 3 Text
	Description string
	NewEntry    bool
}

func (ts TextState) String() string {
	stateText := strings.Join(ts.StateText, ", ")
	return fmt.Sprintf("%s,%s,%d,[%s]", ts.Description, ts.TableName, ts.Index, stateText)
}

var SDstateTextTable = []TextState{
	{TableName: "SD State Text Table", Index: 0, StateText: []string{"RESET", "OPERATED", "", ""}, Description: "Texts for Trip Relay Reset Part"},
	{TableName: "SD State Text Table", Index: 1, StateText: []string{"OFF", "OPERATED", "", ""}, Description: "Texts for Audible Warning Al"},
	{TableName: "SD State Text Table", Index: 2, StateText: []string{"ON", "OFF", "Unsolicited ON", "Unsolicited OFF"}, Description: "Texts for OFF/ON SD Prot"},
	{TableName: "SD State Text Table", Index: 3, StateText: []string{"IN", "OUT", "Unsolicited IN", "Unsolicited OUT"}, Description: "Texts for In/OUT SD Prot"},
	{TableName: "SD State Text Table", Index: 4, StateText: []string{"NORMAL", "OPERATED", "NORMAL", "OPERATED"}, Description: "Texts for RING CHECK Prot"},
	{TableName: "SD State Text Table", Index: 5, StateText: []string{"OFF", "ON", "Unsolicited OFF", "Unsolicited ON"}, Description: "Texts for VR Controls"},
}

var DDStateTextTable = []TextState{
	{TableName: "DD State Text Table", Index: 0, StateText: []string{"DBI-00", "OPEN", "CLOSED", "DBI-11", "DBI-00", "TRIPPED", "Unsolicited CLOSED", "DMI-11"}, Description: "Texts for Typical CB"},
	{TableName: "DD State Text Table", Index: 1, StateText: []string{"DBI-00", "OPEN", "CLOSED", "DMI-11", "DBI-00", "Unsolicited OPEN", "Unsolicited CLOSED", "DMI-11"}, Description: "Texts for Disconnector"},
	{TableName: "DD State Text Table", Index: 2, StateText: []string{"DBI", "OUT", "IN", "DMI", "DBI", "UNSOLICITED OUT", "UNSOLICITED IN", "DMI"}, Description: "Texts for Sec Automation"},
	{TableName: "DD State Text Table", Index: 3, StateText: []string{"DBI", "OFF", "ON", "DMI", "DBI", "UNSOLICITED OFF", "UNSOLICITED ON", "DMI"}, Description: "Text for CNTLs (IEC)"},
	{TableName: "DD State Text Table", Index: 4, StateText: []string{"", "", "", "", "", "", "", ""}, Description: ""},
	{TableName: "DD State Text Table", Index: 5, StateText: []string{"", "", "", "", "", "", "", ""}, Description: ""},
	{TableName: "DD State Text Table", Index: 6, StateText: []string{"", "", "", "", "", "", "", ""}, Description: ""},
	{TableName: "DD State Text Table", Index: 7, StateText: []string{"", "", "", "", "", "", "", ""}, Description: ""},
	{TableName: "DD State Text Table", Index: 8, StateText: []string{"DBI", "NOT AVAIL", "AVAIL", "DMI", "DBI", "UNSOLICITED OUT", "UNSOLICITED IN", "DMI"}, Description: "Texts for SPT Gen Prot"},
}

var SDalarmStateTextTable = []TextState{
	{TableName: "SD Alarm State Text Table", Index: 0, StateText: []string{"NORMAL", "OPERATED"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 1, StateText: []string{"NORMAL", "ALARM"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 2, StateText: []string{"MAIN", "STANDBY"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 3, StateText: []string{"NORMAL", "LOW"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 4, StateText: []string{"NORMAL", "FAULTY"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 5, StateText: []string{"CHARGED", "DISCHARGED"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 6, StateText: []string{"OFF", "ON"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 7, StateText: []string{"IN", "OUT"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 8, StateText: []string{"SUPERVISORY", "LOCAL"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 9, StateText: []string{"AUTO", "MANUAL"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 10, StateText: []string{"AUTO RECLOSE", "AUTO CLOSE"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 11, StateText: []string{"NORMAL", "OPERATED/INHIBIT"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 12, StateText: []string{"IN SERVICE", "OUT OF SERVICE"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 13, StateText: []string{"SWITCHED IN", "SWITCHED OUT"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 14, StateText: []string{"NORMAL", "SPLIT"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 15, StateText: []string{"NORMAL", "INITIATED"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 16, StateText: []string{"DISABLED", "ENABLED"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 17, StateText: []string{"AUTO", "NON-AUTO"}, Description: "Texts for Single Digital Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 18, StateText: []string{"NORMAL", "COMPLETE"}, Description: "Texts for Auto Scheme Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 19, StateText: []string{"NORMAL", "FAILED"}, Description: "Texts for Auto Scheme Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 20, StateText: []string{"NORMAL", "OPERATING"}, Description: "Texts for Auto Scheme Alarms"},
	{TableName: "SD Alarm State Text Table", Index: 21, StateText: []string{"", ""}, Description: ""},
	{TableName: "SD Alarm State Text Table", Index: 22, StateText: []string{"", ""}, Description: ""},
	{TableName: "SD Alarm State Text Table", Index: 23, StateText: []string{"", ""}, Description: ""},
	{TableName: "SD Alarm State Text Table", Index: 24, StateText: []string{"", ""}, Description: ""},
	{TableName: "SD Alarm State Text Table", Index: 25, StateText: []string{"OFF", "ON"}, Description: "Texts for OFF / ON SD Protection Controls. OFF = ABNORMAL = Scanned Value 0"},
	{TableName: "SD Alarm State Text Table", Index: 26, StateText: []string{"OUT", "IN"}, Description: "Texts for In ? OUT SD DAR Protection Controls. OFF = ABRMORMAL = Scanned Value 0"},
}

var DDalarmStateTextTable = []TextState{
	{TableName: "DD Alarm State Text Table", Index: 0, StateText: []string{"DBI", "LOCAL", "SUPERVISORY", "DMI"}, Description: "Texts for DD Automation and"},
	{TableName: "DD Alarm State Text Table", Index: 1, StateText: []string{"NORMAL", "FAILING", "LOW", "LOW"}, Description: "Texts for DD Gas Alarms"},
}

type TextStateTable struct {
	StateTextTable string
	TextState      []TextState
	PointType      string
}

var StateTextTables = map[string]TextStateTable{
	"SD State Text Table":       {StateTextTable: "SD State Text Table", TextState: SDstateTextTable, PointType: "SD"},
	"DD State Text Table":       {StateTextTable: "DD State Text Table", TextState: DDStateTextTable, PointType: "DD"},
	"SD Alarm State Text Table": {StateTextTable: "SD Alarm State Text Table", TextState: SDalarmStateTextTable, PointType: "SD"},
	"DD Alarm State Text Table": {StateTextTable: "DD Alarm State Text Table", TextState: DDalarmStateTextTable, PointType: "DD"},
}

func GetTextState(stateTextTable string, index int, pointType string) (TextState, error) {
	if state, ok := StateTextTables[stateTextTable]; ok {
		if state.PointType != pointType {
			return TextState{}, fmt.Errorf("point type mismatch")
		}

		if index >= len(state.TextState) {
			return TextState{}, fmt.Errorf("%d, %s: index out of range for state text table", index, stateTextTable)
		}
		return state.TextState[index], nil
	}

	return TextState{}, fmt.Errorf("%s : state text table not found", stateTextTable)
}

func (ts TextState) GetTextIndex(text string) (int, error) {
	for index, stateText := range ts.StateText {
		if stateText == text {
			return index, nil
		}
	}
	return -1, fmt.Errorf("%s, %s :text not found in state text table", text, ts.TableName)
}

// func (tst TextStateTable) MatchStateText(ts *TextState) (int, error) {

// 	table :=
