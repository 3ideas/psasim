package compare

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/3ideas/psasim/lib/alarmstatetext"
	"github.com/3ideas/psasim/lib/comps"
	"github.com/3ideas/psasim/lib/namerif"
	"github.com/gocarina/gocsv"
	"github.com/tealeg/xlsx"
)

type AlarmCompare struct {
	RTU_Name               string `csv:"RTU_Name" xlsx:"0"`
	RTU_Address            string `csv:"RTU_Address" xlsx:"1"`
	ETerraAlias            string `csv:"eTerra Alias" xlsx:"2"`
	POAlias                string `csv:"PO Alias" xlsx:"3"`
	Type                   string `csv:"Type" xlsx:"4"`
	Card                   string `csv:"Card" xlsx:"5"`
	Offset                 string `csv:"Offset" xlsx:"6"`
	Value                  string `csv:"Value" xlsx:"7"`
	ETerraSubstation       string `csv:"eTerraSubstation" xlsx:"8"`
	ETerraAlarmMessage     string `csv:"eTerraAlarmMessage" xlsx:"9"`
	ETerraAlarmZone        string `csv:"eTerraAlarmZone" xlsx:"10"`
	ETerraStatus           string `csv:"eTerraStatus" xlsx:"11"`
	POSubstation           string `csv:"POSubstation" xlsx:"12"`
	POAlarmMessage         string `csv:"POAlarmMessage" xlsx:"13"`
	POAlarmZone            string `csv:"POAlarmZone" xlsx:"14"`
	POAlarmValue           string `csv:"POAlarmValue" xlsx:"15"`
	POAlarmRef             string `csv:"POAlarmRef" xlsx:"16"`
	POStatus               string `csv:"POStatus" xlsx:"17"`
	EventCategory          string `csv:"EventCategory" xlsx:"18"`
	DevID                  string `csv:"DevID" xlsx:"19"`
	PointID                string `csv:"PointID" xlsx:"20"`
	AlarmType              string `csv:"AlarmType" xlsx:"21"`
	EToken1                string `csv:"etoken1" xlsx:"22"`
	EToken2                string `csv:"etoken2" xlsx:"23"`
	EToken3                string `csv:"etoken3" xlsx:"24"`
	EToken4                string `csv:"etoken4" xlsx:"25"`
	EToken5                string `csv:"etoken5" xlsx:"26"`
	PToken1                string `csv:"ptoken1" xlsx:"27"`
	PToken2                string `csv:"ptoken2" xlsx:"28"`
	PToken3                string `csv:"ptoken3" xlsx:"29"`
	PToken4                string `csv:"ptoken4" xlsx:"30"`
	PToken5                string `csv:"ptoken5" xlsx:"31"`
	T1Match                string `csv:"T1Match" xlsx:"32"`
	T2Match                string `csv:"T2Match" xlsx:"33"`
	T3Match                string `csv:"T3Match" xlsx:"34"`
	T4Match                string `csv:"T4Match" xlsx:"35"`
	T5Match                string `csv:"T5Match" xlsx:"36"`
	MatchScore             string `csv:"MatchScore" xlsx:"37"`
	AlarmMessageMatch      string `csv:"AlarmMessageMatch" xlsx:"38"`
	AlarmZoneMatch         string `csv:"AlarmZoneMatch" xlsx:"39"`
	Action                 string `csv:"Action"`
	NodeType               string `csv:"NodeType"`
	Spath                  string `csv:"Spath"`
	ParentsAllMatch        bool   `csv:"ParentsMatch"`
	ComponentClass         string `csv:"ComponentClass"`
	SubstationClass        string `csv:"SubstationClass"`
	OriginalAlarmMatch     bool   `csv:"OriginalAlarmMatch"` // in the hierarchy file A = "0" indicate that alarm didn't match 1 indicate that alarm matched this was mapped to the NeedsFixed field.
	E3record               string `csv:"e3record"`
	EterraPrimaryCircuit   string `csv:"eTerraPrimaryCircuit"`
	POPrimaryCircit        string `csv:"POPrimaryCircit"`
	PrevAlarmMessageMatch  string `csv:"PrevAlarmMessageMatch"`
	PrevPOAlarmMessage     string `csv:"PrevPOAlarmMessage"`
	PrevPToken1            string `csv:"PrevPToken1"`
	PrevPToken2            string `csv:"PrevPToken2"`
	PrevPToken3            string `csv:"PrevPToken3"`
	PrevPToken4            string `csv:"PrevPToken4"`
	PrevPToken5            string `csv:"PrevPToken5"`
	PrevAction             string `csv:"PrevAction"`
	AlarmSameAsPreviousRun bool   `csv:"AlarmSameAsPreviousRun"`
	ActionChanged          bool   `csv:"ActionChanged"`
	ComponentTemplateID    string `csv:"ComponentTemplateID"`
	ComponentTemplateName  string `csv:"ComponentTemplateName"`
	ComponentNameRule      string `csv:"ComponentNameRule"`
	*alarmstatetext.ScadaAttributes
}

func readAlarmListFromCSV(filename string) ([]*AlarmCompare, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	alarms := make([]*AlarmCompare, 0)
	if err := gocsv.UnmarshalFile(file, &alarms); err != nil {
		return nil, err
	}

	return alarms, nil

}

func readAlarmListFromXLSX(filename string) ([]*AlarmCompare, error) {
	var alarms []*AlarmCompare

	// Open the XLSX file
	xlFile, err := xlsx.OpenFile(filename)
	if err != nil {
		return nil, err
	}

	// Get the first sheet
	sheet := xlFile.Sheets[0]

	// Iterate over the rows in the sheet
	first := true
	for _, row := range sheet.Rows {
		var alarm AlarmCompare
		err := row.ReadStruct(&alarm)
		if err != nil {
			return nil, err
		}
		if first {
			first = false
			continue
		}
		alarms = append(alarms, &alarm)
	}

	return alarms, nil
}

type AlarmsComparison struct {
	alarms []*AlarmCompare

	alarmsByAliasAlarm map[string]*AlarmCompare
	alarmsByAliasPO    map[string][]*AlarmCompare
}

// All alarms
func (a *AlarmsComparison) Alarms() []*AlarmCompare {
	return a.alarms
}

func NewAlarmsComparison() *AlarmsComparison {
	return &AlarmsComparison{
		alarms:             make([]*AlarmCompare, 0),
		alarmsByAliasAlarm: make(map[string]*AlarmCompare),
		alarmsByAliasPO:    make(map[string][]*AlarmCompare),
	}
}

func (a *AlarmsComparison) AddAlarm(alarm *AlarmCompare) {
	a.alarms = append(a.alarms, alarm)
	key := alarm.POAlias + alarm.ETerraAlarmMessage
	a.alarmsByAliasAlarm[key] = alarm

	alarmList, ok := a.alarmsByAliasPO[alarm.POAlias]
	if !ok {
		alarmList = make([]*AlarmCompare, 0)
	}
	alarmList = append(alarmList, alarm)
	a.alarmsByAliasPO[alarm.POAlias] = alarmList
}

func (a *AlarmsComparison) GetAlarmsByAliasPO(alias string) []*AlarmCompare {
	return a.alarmsByAliasPO[alias]
}

func (a *AlarmsComparison) Match(alarm *AlarmCompare) (*AlarmCompare, bool) {
	alarm2, ok := a.alarmsByAliasAlarm[alarm.POAlias+alarm.ETerraAlarmMessage]

	return alarm2, ok
}

func (a *AlarmsComparison) checkIfParentsOK(comp *comps.Component) bool {

	parent := comp.Parent
	for parent != nil {
		alarms := a.GetAlarmsByAliasPO(parent.Alias)
		if alarms == nil {
			parent = parent.Parent
			continue
		}
		for _, alarm := range alarms {
			if alarm.AlarmMessageMatch == "FALSE" {
				return false
			}
		}
		parent = parent.Parent
	}
	return true
}

func (a *AlarmsComparison) AddActionsAndInfo(comps *comps.ComponentManager, nameserver namerif.NameService) {

	stateFixup := NewStateFixup()
	for _, alarm := range a.alarms {
		comp, ok := comps.GetCompByAlias(alarm.POAlias)
		if !ok {
			// fmt.Printf("Could not find component for alias in hierarchy %s\n", alarm.POAlias)
			slog.Info("AddActions: Could not find component for alias in hierarchy, skipping", "alias", alarm.POAlias)
			continue
		}

		cloneId := ""
		clonePathname := ""
		nameRule := ""

		if nameserver != nil {
			compInfo, err := nameserver.GetComponentInfo(comp.Alias)
			if err == nil {
				cloneId = compInfo.CloneID
				clonePathname = compInfo.ClonePathname
				nameRule = compInfo.NameRule
			}
			scadaAttr, err := alarmstatetext.GetScadaAttributes(comp.Alias, nameserver, alarm.PointID, alarm.Value, alarm.EToken5, alarm.PToken5, alarm.Type)
			if err == nil {
				alarm.ScadaAttributes = scadaAttr

				if alarm.T5Match == "0" {
					if alarm.POAlias == "BUSB2/275_LN/BUSB_STHA/B003" {
						fmt.Printf("Adding state text for alias: %s, Table: %s, PointID: %s, eToken5: %s, StateTextIndex: %d\n", alarm.POAlias, scadaAttr.StateTextTable, scadaAttr.PointID, alarm.EToken5, scadaAttr.StateTextIndex)
					}
					err = stateFixup.AddPointIDTextState(scadaAttr.StateTextTable, scadaAttr.StateTextIndex, scadaAttr.PointID, alarm.EToken5, alarm.POAlias)
					if err != nil {
						scadaAttr.FixupError = err.Error()
						fmt.Printf("AddActionsAndInfo: Error adding state text. Alias: %s, Table: %s, PointID: %s, eToken5: %s, Error: %s\n", alarm.POAlias, scadaAttr.StateTextTable, scadaAttr.PointID, alarm.EToken5, err)
						slog.Error("AddActionsAndInfo: Error adding state text", "error", err, "alias", alarm.POAlias, "table", scadaAttr.StateTextTable, "pointID", scadaAttr.PointID, "eToken5", alarm.EToken5)
					}
				}
			}
		}

		alarm.Action = comp.Action
		if len(comp.Children) == 0 {
			alarm.NodeType = "Leaf"
		} else {
			alarm.NodeType = "Branch"
		}
		// alarm.Spath = comp.SPATH

		alarm.ParentsAllMatch = a.checkIfParentsOK(comp)

		alarm.ComponentClass = comp.ComponentClass
		alarm.SubstationClass = comp.SubstationClass
		// alarm.OriginalAlarmMatch = !comp.AlarmNeedsFix // in the hierarchy file A = "0" indicate that alarm didn't match 1 indicate that alarm matched this was mapped to the NeedsFixed field.
		alarm.E3record = comp.EterraToken3
		// alarm.EterraPrimaryCircuit = comp.EterraPrimaryCircuit
		// alarm.POPrimaryCircit = comp.PoPrimaryCircuit
		alarm.ComponentTemplateID = cloneId
		alarm.ComponentTemplateName = clonePathname
		alarm.ComponentNameRule = nameRule

	}

	stateFixup.DisplayAll()
	// stateFixup.FixupAll()
}

func Compare(a, b *AlarmCompare) bool {
	return a.POAlarmMessage == b.POAlarmMessage
}

func ReadAlarmComparison(filename string) (*AlarmsComparison, error) {
	var alarms []*AlarmCompare
	var err error

	if filename[len(filename)-5:] == ".xlsx" {
		alarms, err = readAlarmListFromXLSX(filename)
	} else if filename[len(filename)-4:] == ".csv" {
		alarms, err = readAlarmListFromCSV(filename)
	} else {
		err = fmt.Errorf("unknown file type: %s", filename)
	}

	if err != nil {
		return nil, err
	}

	alarmList := NewAlarmsComparison()
	for _, alarm := range alarms {
		alarmList.AddAlarm(alarm)
	}

	return alarmList, nil
}

func (a *AlarmsComparison) Merge(b *AlarmsComparison) {

	for _, alarm := range a.alarms {

		alarm2, ok := b.Match(alarm)
		if !ok {
			continue
		}
		alarm.PrevAlarmMessageMatch = alarm2.AlarmMessageMatch
		alarm.PrevPOAlarmMessage = alarm2.POAlarmMessage
		alarm.PrevPToken1 = alarm2.PToken1
		alarm.PrevPToken2 = alarm2.PToken2
		alarm.PrevPToken3 = alarm2.PToken3
		alarm.PrevPToken4 = alarm2.PToken4
		alarm.PrevPToken5 = alarm2.PToken5
		alarm.PrevAction = alarm2.Action
		alarm.AlarmSameAsPreviousRun = Compare(alarm, alarm2)

		if alarm.Action != alarm.PrevAction {
			alarm.ActionChanged = true
		}
	}
}

func (a *AlarmsComparison) WriteToCSV(filename string) error {

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	csvContent, err := gocsv.MarshalString(&a.alarms)
	if err != nil {
		return err
	}
	f.WriteString(csvContent)
	return nil
}

func (a *AlarmsComparison) WriteWorkingActions(filename string) error {

	actions := NewActions()

	for _, alarm := range a.alarms {
		if _, ok := actions.GetAction(alarm.POAlias); ok {
			continue
		}
		if alarm.AlarmMessageMatch == "FALSE" {
			continue
		}

		action := &Action{
			POAlias:      alarm.POAlias,
			Action:       alarm.Action,
			ParentsMatch: alarm.ParentsAllMatch,
		}
		actions.AddAction(action)
	}

	return actions.WriteToCSV(filename)
}
