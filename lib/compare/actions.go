package compare

import (
	"os"

	"github.com/gocarina/gocsv"
)

type Action struct {
	POAlias      string `csv:"Alias" `
	Action       string `csv:"Action" `
	ParentsMatch bool   `csv:"ParentsMatch" `
}

type Actions struct {
	Actions       []*Action
	aliasToAction map[string]*Action
}

func NewActions() *Actions {
	return &Actions{
		Actions:       make([]*Action, 0),
		aliasToAction: make(map[string]*Action),
	}
}

func (a *Actions) AddAction(action *Action) {
	a.Actions = append(a.Actions, action)
	a.aliasToAction[action.POAlias] = action
}

func (a *Actions) GetAction(alias string) (*Action, bool) {
	action, ok := a.aliasToAction[alias]
	return action, ok
}

func ReadActions(filename string) (*Actions, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	actions := make([]*Action, 0)
	if err := gocsv.UnmarshalFile(file, &actions); err != nil {
		return nil, err
	}
	actionsList := NewActions()
	for _, action := range actions {
		actionsList.AddAction(action)
	}
	return actionsList, nil
}

func (a *Actions) WriteToCSV(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	csvContent, err := gocsv.MarshalString(&a.Actions)
	if err != nil {
		return err
	}
	f.WriteString(csvContent)
	return nil
}
