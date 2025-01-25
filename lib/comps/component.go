package comps

import (
	"regexp"
	"sort"
	"strings"
)

// A,T,Path / Combined Alarm Message,Alias / eTerra Alarm Message,Name / PO Alarm Message,Circuit,e3/possible correct circuit,Substation Class,Component Class

type Component struct {
	Path            string
	Alias           string
	Name            string
	Parent          *Component
	Children        []*Component
	CurrentCircuit  string
	EterraToken3    string
	SubstationClass string
	ComponentClass  string
	// EterraPrimaryCircuit string
	PoCircuit      string
	Action         string
	OriginalParent *Component // If we move a component then this will hold the original parent otherwise it will be nil
}

func findAllSubstringsInBraces(s string) []string {
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindAllStringSubmatch(s, -1)
	var results []string
	for _, match := range matches {
		results = append(results, match[1])
	}
	return results
}

// func (c *Component) AddAlarmMsg(a *Alarm) {

// 	// From the CombinedAlarmMsg extract the embedded field enclised in []
// 	pc := findAllSubstringsInBraces(a.CombinedAlarmMsg)
// 	if len(pc) > 0 { // Found at least one set of fields // always use the first one
// 		fieldNumber := 0
// 		// if the CombinedAlarmMsg starts with a [  then we need to use the 2nd field, See Alan on details why this is. This is because someetimes the substation name does not match then the circuit so it is not the firstbraced field.
// 		if strings.HasPrefix(a.CombinedAlarmMsg, "[") {
// 			fieldNumber = 1
// 			if len(pc) < 2 {
// 				fieldNumber = -1 // a.ErrorProcessing = "CombinedAlarmMsg has no fields in []"
// 				return
// 			}
// 		}

// 		if fieldNumber >= 0 {
// 			fields := strings.Split(pc[fieldNumber], "|")
// 			if len(fields) == 2 {

// 				// a.EterraPrimaryCircuit = fields[0]
// 				// a.PoPrimaryCircuit = fields[1]
// 				c.EterraPrimaryCircuit = fields[0]
// 				c.PoCircuit = fields[1]
// 			}
// 		}
// 	} else {

// 		a.ErrorProcessing = "CombinedAlarmMsg has no fields in []"
// 	}
// 	// if a.A == "0" {
// 	// 	c.AlarmNeedsFix = true
// 	// }

// }

// HasAnyParentChanged checks if any parent has been changed, this is used
// to determine if a component needs to be checked for alarm changes
// This is to deal with that alarms that where originally marked as not needing fixing may need fixed if
// a parent has been changed.
func (c *Component) HasAnyParentChanged() bool {
	comp := c
	for comp != nil {
		if comp.OriginalParent != nil {
			return true
		}
		comp = comp.Parent
	}
	return false
}

// func (c *Component) BuildSpath(assignToOriginal bool) {
// 	letter := "aA"
// 	for _, child := range c.Children {
// 		leafFlag := ""
// 		if len(child.Children) == 0 {
// 			leafFlag = "+"
// 		}
// 		child.SPATH = c.SPATH + letter + leafFlag

// 		child.BuildSpath(assignToOriginal)
// 		// incremnt the letter
// 		letter = NextLetter(letter)
// 	}
// }

// RemoveChild removes a child from the component
func (c *Component) RemoveChild(child *Component) {
	for i, comp := range c.Children {
		if comp == child {
			c.Children = append(c.Children[:i], c.Children[i+1:]...)
			break
		}
	}
}

func (c *Component) IsLeafNode() bool {
	return len(c.Children) == 0
}

// CheckIfChild checks if a child is a child of the component anywhere in its children..
func (c *Component) CheckIfChild(child *Component) bool {
	for _, comp := range c.Children {
		if comp == child {
			return true
		}
		if comp.CheckIfChild(child) {
			return true
		}
	}
	return false
}

// AddChild adds a child to the component
func (c *Component) AddChild(child *Component) {
	child.Parent = c
	c.Children = append(c.Children, child)
}

func (c *Component) IsSwitch() bool {
	return strings.HasSuffix(c.Alias, "SWDD")
}

func (c *Component) IsSGT() bool {
	return strings.Contains(c.Alias, "/SGT")
}

func (c *Component) IsDCB() bool {
	// a DCB will have the string "/DCB/" in the alias
	return strings.Contains(c.Alias, "/DCB/")
}

func (c *Component) IsSCE1() bool {
	return strings.Contains(c.Alias, "/SC1E")
}

func (c *Component) IsSCE2() bool {
	return strings.Contains(c.Alias, "/SC2E")
}

func (c *Component) IsCircuit() bool {
	return c.SubstationClass == "Primary Circuit ID"
}

func (c *Component) SortChildren() {

	sort.Slice(c.Children, func(i, j int) bool {
		return c.Children[i].Alias < c.Children[j].Alias
	})

	for _, child := range c.Children {
		child.SortChildren()
	}
}

func (c *Component) Is25kvCB() bool {
	return strings.Contains(c.Alias, "/025_CB/")
}

func (c *Component) GetPrimaryCircuitComp() *Component {

	comp := c

	for comp != nil && comp.SubstationClass != "Primary Circuit ID" {
		comp = comp.Parent
	}

	return comp
}

// GetGroupingComp returns the grouping component for the component, this is the point under which any circuit or non circuit components will be placed

func (c *Component) GetGroupingComp() *Component {

	comp := c

	// Find the first grouping point
	for comp != nil && !(comp.SubstationClass == "7" || comp.SubstationClass == "Location Holder" || comp.ComponentClass == "701") { // - this is now enumerating at 7 in the file.. but the src file is als0 changing it back again ... added both conditions
		comp = comp.Parent
	}

	if comp == nil { //No grouping point found, try for the primary sub
		comp = c

		for comp != nil && comp.SubstationClass != "Primary Substation" {
			comp = comp.Parent
		}

	}

	return comp
}
