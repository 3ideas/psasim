package comps

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/3ideas/psasim/lib/classification"
)

// func extractField(s string) string {
// 	start := strings.Index(s, "[")
// 	end := strings.Index(s, "]")
// 	if start != -1 && end != -1 && start < end {
// 		return s[start+1 : end]
// 	}
// 	return ""
// }

type ComponentManager struct {
	FileName      string
	Comps         []*Component
	Root          *Component
	rootPath      string
	aliasToComp   map[string]*Component
	pathToComp    map[string]*Component
	rollbackStack []RollbackOperation
	rollbackPoint int
	// *classification.Classifications
}

func NewComponentManager(filename string) *ComponentManager {
	return &ComponentManager{
		FileName:    filename,
		aliasToComp: make(map[string]*Component),
		pathToComp:  make(map[string]*Component),
		// Classifications: classifications,
	}
}

func (c *ComponentManager) AddWithPathMod(comp *Component) error {

	if c.Root == nil { // Assume the first component is the root
		c.Root = comp
		comp.Path = comp.Alias // As this records path is not as expected...
		c.rootPath = comp.Path
	} else {
		comp.Path = c.rootPath + ":" + comp.Path
	}

	return c.Add(comp)
}

func (c *ComponentManager) AddWithoutPathMod(comp *Component) error {

	if c.Root == nil { // Assume the first component is the root
		c.Root = comp
		// comp.Path = comp.Alias // As this records path is not as expected...
		// c.rootPath = comp.Path
	}

	pathParts := strings.Split(comp.Path, ":")

	if pathParts[0] == "ROOT" {
		// This has the full root path. remove the 1sr 3 parts to get to the substation parth
		// TODO: this is a bit of a hack as the inpuput file should not have it like this!!
		comp.Path = strings.Join(pathParts[3:len(pathParts)], ":")
	}
	return c.Add(comp)
}

// Add adds a component to the list of components, it assumues the component has a parent and the parent already exisits.
func (c *ComponentManager) Add(comp *Component) error {

	if _, ok := c.pathToComp[comp.Path]; ok {
		fmt.Printf("Path duplicated when adding component: %s, %s \n", comp.Alias, comp.Path)
		return fmt.Errorf("Component already added. Path dumplcation detected: %s, %s ", comp.Alias, comp.Path)
	}

	c.Comps = append(c.Comps, comp)

	c.aliasToComp[comp.Alias] = comp
	c.pathToComp[comp.Path] = comp

	if c.Root == nil {
		c.Root = comp
		return nil
	}

	parent := comp.Parent
	if parent != nil {
		// Don't change the parenting, but add to the children list
		parent.Children = append(parent.Children, comp)
		return nil
	}

	// Find the parent of the component
	parentPath := getParentPath(comp.Path)
	parent, ok := c.pathToComp[parentPath]
	if ok {
		// Add children to parent
		parent.Children = append(parent.Children, comp)
		comp.Parent = parent
	}
	return nil
}

func (c *ComponentManager) RemoveComponent(alias string) error {

	comp, ok := c.GetCompByAlias(alias)
	if !ok {
		return fmt.Errorf("no such component: %s", alias)
	}

	comp.Parent.RemoveChild(comp)

	delete(c.aliasToComp, alias)
	delete(c.pathToComp, comp.Path)

	// Horrible
	for i, c2 := range c.Comps {
		if c2 == comp {
			c.Comps = append(c.Comps[:i], c.Comps[i+1:]...)
			break
		}
	}

	return nil
}

// func (c *ComponentManager) BuildSPaths(initalLetter string, assignToOriginal bool) {

// 	comp := c.Root
// 	if initalLetter == "" {
// 		initalLetter = comp.Path
// 	}
// 	comp.SPATH = initalLetter

// 	comp.BuildSpath(assignToOriginal)
// }

func incLetter(letter int32) rune {

	r := rune(letter)
	return r + 1
}

func NextLetter(letter string) string {
	if len(letter) > 2 {
		return letter // Return the original letter if it's not a single character
	}
	prefix := rune(letter[0])
	extension := rune(letter[1])
	if prefix == 'z' {
		extension = incLetter(extension)
		prefix = 'a'
		return string(prefix) + string(extension)
	}
	prefix = incLetter(prefix)
	return string(prefix) + string(extension)
}

func (c *ComponentManager) GetCompByAlias(alias string) (*Component, bool) {
	comp, ok := c.aliasToComp[alias]
	return comp, ok
}

func (c *ComponentManager) GetCompByPath(path string) (*Component, bool) {
	comp, ok := c.pathToComp[path]
	return comp, ok
}

func (c *ComponentManager) IsCompPathPresent(path string) bool {
	_, ok := c.pathToComp[path]
	return ok
}

func (c *ComponentManager) IsCompAliasPresent(alias string) bool {
	_, ok := c.aliasToComp[alias]
	return ok
}

func ReadComps(filename string, origFileFormat bool, classifications *classification.Classifications, pathLetter string, compsOfInterest map[string]struct{}) (*ComponentManager, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the csv file
	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// var currentComp *Component

	first := true

	comps := NewComponentManager(filename)

	lineNo := 0
	for _, record := range records {
		lineNo++
		if first {
			first = false
			continue
		}
		if record[1] != "" && record[1] != " " { // Its a path record
			if len(record[2]) > 128 {
				return nil, fmt.Errorf("line: %d, alias: %s, is too long", lineNo, record[2])
			}
			comp := &Component{Path: record[2], Alias: record[3], Name: record[4], CurrentCircuit: record[5], EterraToken3: record[6], SubstationClass: record[7], ComponentClass: record[8]}

			var err error
			if origFileFormat {
				err = comps.AddWithPathMod(comp)
			} else {
				err = comps.AddWithoutPathMod(comp)
			}

			if err != nil {
				fmt.Printf("Error adding component: %s\n", err)
				continue
			}

			// currentComp = comp
		} else {
			// Its an alarm record

			// alarm := &Alarm{
			// 	A:                record[0],
			// 	T:                record[1],
			// 	CombinedAlarmMsg: record[2],
			// 	ETerraAlarmMsg:   record[3],
			// 	POAlarmMsg:       record[4],
			// 	Circuit:          record[5],
			// 	e3Circuit:        record[6],
			// 	SubstationClass:  record[7],
			// 	ComponentClass:   record[8],
			// 	LineNo:           lineNo,
			// }
			// if _, ok := compsOfInterest[currentComp.Alias]; ok {
			// 	fmt.Printf("Comp of interest: %s \n", currentComp.Alias)
			// }

			// currentComp.AddAlarmMsg(alarm)
		}
	}

	// If we only have 1 component, this was an empty file with just the substation, return an error
	if len(comps.Comps) == 1 {
		return nil, fmt.Errorf("no components found in file")
	}

	comps.Root.SortChildren()
	// comps.BuildSPaths(pathLetter, true)

	return comps, nil

}

func GenBaseComponents(classifications *classification.Classifications) *ComponentManager {
	comps := NewComponentManager("")
	comp := &Component{Path: "ROOT", Alias: "ROOT", Name: "ROOT"}
	err := comps.Add(comp)
	if err != nil {
		fmt.Printf("Error adding BASE component: %s\n", err)
		return nil
	}
	return comps
}

func (comps *ComponentManager) AddSet(newComps *ComponentManager) error {

	newComps.Root.Parent = comps.Root

	for _, comp := range newComps.Comps {
		err := comps.Add(comp)
		if err != nil {
			return err
		}
	}
	return nil
}

func getParentPath(path string) string {

	pathParts := strings.Split(path, ":")
	if len(pathParts) == 1 {
		return ""
	}

	var parentPath string
	if pathParts[0] == "ROOT" {
		// This has the full root path. remove the 1sr 3 parts to get to the substation parth
		// TODO: this is a bit of a hack as the inpuput file should not have it like this!!
		parentPath = strings.Join(pathParts[3:len(pathParts)-1], ":")
	} else {
		// Joint the parts except the last one
		parentPath = strings.Join(pathParts[:len(pathParts)-1], ":")
	}
	return parentPath
}

func (c *Component) GetPath() string {
	names := []string{}
	comp := c
	for comp != nil {
		names = append(names, comp.Name)
		comp = comp.Parent
	}
	// Reverse the names
	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}
	return strings.Join(names, ":")
}

// func (c *Component) GetCircuitName() string {
// 	comp := c

// 	for comp != nil {
// 		if comp.CurrentCircuit != "" {
// 			return comp.CurrentCircuit
// 		}
// 		comp = comp.Parent
// 	}

// 	return ""
// }
