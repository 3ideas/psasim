package compdb

import (
	"fmt"
	"os"

	_ "github.com/glebarez/go-sqlite"
	"github.com/jmoiron/sqlx"
)

type ComponentDb struct {
	db *sqlx.DB
	*ComponentClassDefns
	*ComponentNameRules
	*Components
	*Attributes

	rollbackPoint int
	rollbackStack []RollbackOperation
}

func NewCompDb() *ComponentDb {
	return &ComponentDb{
		rollbackStack:       []RollbackOperation{},
		ComponentClassDefns: NewComponentClassDefns(),
		ComponentNameRules:  NewComponentNameRules(),
		Components:          NewComponentManager(),
		Attributes:          NewAttributeManager(),
	}
}

func LoadCompDb(dbFile string) (*ComponentDb, error) {

	var namer ComponentDb

	db, err := OpenDB(dbFile)
	if err != nil {
		return nil, err
	}
	namer.db = db

	namer.ComponentClassDefns, err = GetComponentClasses(db)
	if err != nil {
		return nil, err
	}
	SetComponentClassDefinitions(namer.ComponentClassDefns)

	namer.ComponentNameRules, err = GetComponentNameRules(db)
	if err != nil {
		return nil, err
	}

	namer.Components, err = GetComponents(db)
	if err != nil {
		return nil, err
	}

	namer.Components.BuildHierarchy()

	attributeNames := []string{"State Alarm Text", "Not Valid", "Location Name", "Location ID", "Device Name", "Circuit Name", "Switch Number", "Plant", "Supplementary Text", "State Alarm", "State Index", "Alarm Treatment",
		"State 0 Text", "State 1 Text", "State 2 Text", "State 3 Text", "State 4 Text", "State 5 Text", "State 6 Text", "State 7 Text",
		"State 0 text", "State 1 text", "State 2 text", "State 3 text", "State 4 text", "State 5 text", "State 6 text", "State 7 text",
	}
	namer.Attributes, err = GetAttributes(db, attributeNames)
	if err != nil {
		return nil, err
	}

	fmt.Println("Resolving names")
	namer.ResolveNames()
	fmt.Println("Names resolved")

	return &namer, nil
}

// open the database, retuens the sqlx db object
func OpenDB(dbFile string) (*sqlx.DB, error) {

	readonly := fmt.Sprintf("file:%s?cache=private&mode=ro", dbFile)
	db, err := sqlx.Open("sqlite", readonly)
	if err != nil {
		return nil, err
	}

	// _, err = db.Exec("PRAGMA page_size = 4096")
	// if err != nil {
	// 	return nil, fmt.Errorf("does the db exist? Error setting page_size: %s %v", dbFile, err)
	// }

	// _, err = db.Exec("PRAGMA cache_size=-20000")
	// if err != nil {
	// 	return nil, fmt.Errorf("does the db exist? Error setting cache_size: %s %v", dbFile, err)
	// }
	// _, err = db.Exec("PRAGMA mmap_size=2147483648")
	// if err != nil {
	// 	panic(err)
	// }
	// _, err = db.Exec("PRAGMA page_size = 8192")
	// if err != nil {
	// 	panic(err)
	// }

	return db, nil
}

func (n *ComponentDb) GetNameRulesForComponent(alias string) (string, []*ComponentNameRule, error) {

	if n == nil {
		return "", nil, fmt.Errorf("GetNameRulesForComponent: Namer is nil")
	}

	comp, err := n.GetComponent(alias)
	if err != nil {
		return "", nil, fmt.Errorf("no component found for alias %s, %w", alias, err)
	}

	classDefn, err := n.GetComponentClassDefnByIndex(comp.ComponentClass)
	if err != nil {
		return "", nil, fmt.Errorf("no component class definition found for component %s, %w", alias, err)
	}

	if classDefn.ComponentNameRule == "" {
		return "<default>", DefaultNameRules(), nil
	}

	rules, ok := n.GetComponentNameRule(classDefn.ComponentNameRule)
	if !ok {
		return classDefn.ComponentNameRule, nil, fmt.Errorf("no name rules found for component: %s, class definition: %s, namerule: '%s'", alias, classDefn.ComponentClassName, classDefn.ComponentNameRule)
	}

	return classDefn.ComponentNameRule, rules, nil

}

// GetAttributeValue returns the value of the attribute as a AttributeValue struct this is for the NameService interface
func (n *ComponentDb) GetAttributeValue(alias string, attributeName string) (AttributeValue, error) {

	comp, err := n.GetComponent(alias)
	if err != nil {
		return AttributeValue{}, err
	}

	attr, err := n.GetComponentAttribute(comp.ComponentID, attributeName)
	if err != nil {
		return AttributeValue{}, err
	}
	return AttributeValue{Name: attr.AttributeName, Value: attr.AttributeValue, ID: attr.AttributeID, CompID: attr.ComponentID, Definition: attr.AttributeDefinition}, nil
}

func (n *ComponentDb) DumpNames(filename string) error {

	// Open the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the names to the file
	for _, component := range n.Components.componentsByName {
		for _, comp := range component {
			fmt.Fprintf(file, "%s,%s,\"%s\"\n", comp.ComponentAlias, comp.ComponentPathname, comp.Name)
		}
	}

	return nil
}
