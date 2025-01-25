package compdb

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Component struct {
	ComponentID       string `db:"COMPONENT_ID"`
	ComponentPathname string `db:"COMPONENT_PATHNAME"`
	ComponentAlias    string `db:"COMPONENT_ALIAS"`
	// ComponentVersion          string         `db:"COMPONENT_VERSION"`
	// ComponentLocation         string         `db:"COMPONENT_LOCATION"`
	ComponentParentID string `db:"COMPONENT_PARENT_ID"`
	Parent            *Component
	OriginalParent    *Component
	// ComponentSourceID         string         `db:"COMPONENT_SOURCE_ID"`
	// ComponentDestID           string         `db:"COMPONENT_DEST_ID"`
	// ComponentConnectClass     string         `db:"COMPONENT_CONNECT_CLASS"`
	// ComponentCategories       string         `db:"COMPONENT_CATEGORIES"`
	// ComponentApplicFlags      string         `db:"COMPONENT_APPLIC_FLAGS"`
	// ComponentSwitchStatus     string         `db:"COMPONENT_SWITCH_STATUS"`
	// ComponentStatus           string         `db:"COMPONENT_STATUS"`
	// ProtectionLevel           string         `db:"PROTECTION_LEVEL"`
	// UserReference             string         `db:"USER_REFERENCE"`
	// ComponentType             string         `db:"COMPONENT_TYPE"`
	// ComponentNormalDressing   string         `db:"COMPONENT_NORMAL_DRESSING"`
	// ExternalSource            string         `db:"EXTERNAL_SOURCE"`
	// Naming                    string         `db:"NAMING"`
	// PhasesPresent             string         `db:"PHASES_PRESENT"`
	// PhasesSwitchingMode       string         `db:"PHASES_SWITCHING_MODE"`
	// LocationType              string         `db:"LOCATION_TYPE"`
	// ComponentSLDClass         string         `db:"COMPONENT_SLD_CLASS"`
	ComponentClass           ComponentClassIndex `db:"COMPONENT_CLASS"`
	ComponentSubstationClass SubstationType      `db:"COMPONENT_SUBSTATION_CLASS"`
	ComponentCloneID         string              `db:"COMPONENT_CLONE_ID"`
	// ComponentPatchNumber      int            `db:"COMPONENT_PATCH_NUMBER"`
	// BitSize                   int            `db:"BIT_SIZE"`
	// Easting                   float64        `db:"EASTING"`
	// Northing                  float64        `db:"NORTHING"`
	// ComponentCEEnable         bool           `db:"COMPONENT_CE_ENABLE"`
	// ComponentOwnMaintCtrl     bool           `db:"COMPONENT_OWN_MAINT_CTRL"`
	// ComponentHasCTE           bool           `db:"COMPONENT_HAS_CTE"`
	// ComponentNeedsReplication bool           `db:"COMPONENT_NEEDS_REPLICATION"`
	// AutomaticCircuitNaming    bool           `db:"AUTOMATIC_CIRCUIT_NAMING"`
	// PhasesNormallyOpen        bool           `db:"PHASES_NORMALLY_OPEN"`
	// path     string // TODO: remove this as should use GetPath()
	Children []*Component
}

type Components struct {
	componentsByAlias map[string]*Component
	componentsByID    map[string]*Component
	ByPath            map[string]*Component
	Root              *Component
}

func NewComponentManager() *Components {
	return &Components{
		componentsByAlias: make(map[string]*Component),
		componentsByID:    make(map[string]*Component),
		ByPath:            make(map[string]*Component),
		// childrenByID:      make(map[string][]*Component),
	}
}

func (c *Components) AddComponentNoHierarchy(component *Component) error {
	c.componentsByAlias[component.ComponentAlias] = component
	c.componentsByID[component.ComponentID] = component

	return nil
}

func (c *Components) AddComponent(component *Component) error {
	c.componentsByAlias[component.ComponentAlias] = component
	c.componentsByID[component.ComponentID] = component

	parent, ok := c.componentsByID[component.ComponentParentID]
	if !ok {
		// slog.Error("no parent found for component", "component", component.ComponentAlias)
		return fmt.Errorf("no parent found for component %s", component.ComponentAlias)
	}
	parent.Children = append(parent.Children, component)
	parent.SortChildren()
	component.Parent = parent

	c.ByPath[component.GetFullPath()] = component

	return nil
}

func (c *Components) RemoveComponent(componentID string) error {
	comp, ok := c.componentsByID[componentID]
	if !ok {
		return fmt.Errorf("no component found for ID %s", componentID)
	}

	path := comp.GetFullPath()
	delete(c.componentsByAlias, comp.ComponentAlias)
	delete(c.componentsByID, componentID)
	delete(c.ByPath, path)

	parent, ok := c.componentsByID[comp.ComponentParentID]
	if !ok {
		return fmt.Errorf("no parent found for component %s", comp.ComponentAlias)
	}
	parent.RemoveChild(comp)

	return nil
}

func (c *Component) RemoveChild(child *Component) {
	for i, comp := range c.Children {
		if comp == child {
			c.Children = append(c.Children[:i], c.Children[i+1:]...)
			break
		}
	}
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

func (c *Component) GetNodeType() string {
	if len(c.Children) == 0 {
		return "Leaf"
	} else {
		return "Branch"
	}
}

// GetAllChildernAliases returns all the aliases of the component and all its children. The list is top down (breadth first)
func (c *Component) GetAllChildernAliases() []string {
	aliases := make([]string, 0)
	// Add all the children aliases to the list
	for _, child := range c.Children {
		aliases = append(aliases, child.ComponentAlias)
	}
	// Add all the children aliases of the children to the list
	for _, child := range c.Children {
		childrenAliases := child.GetAllChildernAliases()
		aliases = append(aliases, childrenAliases...)
	}
	return aliases
}

// GetComponent returns the component that matches the given ComponentAlias
// func GetComponent(db *sqlx.DB, componentAlias string) (*Component, error) {
// 	var component Component
// 	err := db.Get(&component, "SELECT COMPONENT_ID,COMPONENT_PATHNAME,COMPONENT_ALIAS,COMPONENT_CLASS,COMPONENT_PARENT_ID  FROM COMPONENT_HEADER WHERE COMPONENT_ALIAS = ? LIMIT 1", componentAlias)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &component, nil
// }

func GetComponents(db *sqlx.DB) (*Components, error) {
	var components []*Component
	err := db.Select(&components, `
	SELECT 
		COMPONENT_ID, 
		COALESCE(COMPONENT_PATHNAME, '') AS COMPONENT_PATHNAME, 
		COMPONENT_ALIAS,
		 COALESCE(COMPONENT_CLASS, 0) AS COMPONENT_CLASS, 
		 COMPONENT_SUBSTATION_CLASS,
		 COALESCE(COMPONENT_PARENT_ID, '') AS COMPONENT_PARENT_ID,
		 COALESCE(COMPONENT_CLONE_ID, 0) AS COMPONENT_CLONE_ID
		 FROM COMPONENT_HEADER WHERE component_patch_number <= 0`)
	if err != nil {
		return nil, err
	}

	compManager := NewComponentManager()
	for _, component := range components {
		compManager.AddComponentNoHierarchy(component)

	}

	return compManager, nil
}

func (c *Components) BuildHierarchy() {
	for _, component := range c.componentsByAlias {
		if component.ComponentParentID == "" {
			c.Root = component
			continue
		}
		parent, ok := c.componentsByID[component.ComponentParentID]
		if !ok {
			slog.Error("no parent found for component", "component", component.ComponentAlias)
			continue
		}
		component.Parent = parent
		parent.Children = append(parent.Children, component)
	}

	// sort children by alias
	for _, comp := range c.componentsByAlias {
		// comp.UpdatePath()
		c.ByPath[comp.GetFullPath()] = comp
		comp.SortChildren()
	}

}

func (c *Component) SortChildren() {
	sort.Slice(c.Children, func(i, j int) bool {
		return c.Children[i].ComponentPathname < c.Children[j].ComponentPathname
	})
}

func (c *Components) GetComponent(componentAlias string) (*Component, error) {
	if c == nil {
		return nil, fmt.Errorf("GetComponent: components is nil. Getting component for alias: %s", componentAlias)
	}
	if c.componentsByAlias == nil {
		return nil, fmt.Errorf("GetComponent: componentsByAlias is nil Getting component for alias: %s", componentAlias)
	}
	comp, ok := c.componentsByAlias[componentAlias]
	if !ok {
		return nil, fmt.Errorf("no component found for alias: %s", componentAlias)
	}
	return comp, nil
}

func (c *Components) GetComponentByID(componentID string) (*Component, error) {
	comp, ok := c.componentsByID[componentID]
	if !ok {
		return nil, fmt.Errorf("no component found for ID %s", componentID)
	}
	return comp, nil
}

func (c *Components) GetComponentByPath(path string) (*Component, bool) {
	comp, ok := c.ByPath[path]
	return comp, ok
}

// func (c *Components) GetChildrenByID(componentID string) ([]*Component, bool) {
// 	children, ok := c.childrenByID[componentID]
// 	return children, ok
// }

func (c *Components) GetParents(alias string) ([]*Component, error) {

	comp, err := c.GetComponent(alias)
	if err != nil {
		return nil, fmt.Errorf("no component found for alias %s, %w", alias, err)
	}

	parents := make([]*Component, 0)

	parentId := comp.ComponentID

	// while parentId is not empty
	for parentId != "" {
		comp, err := c.GetComponentByID(parentId)
		if err != nil {
			return nil, fmt.Errorf("no component found for parent id %s, %w", parentId, err)
		}
		parents = append(parents, comp)
		parentId = comp.ComponentParentID
	}

	return parents, nil

}

// func (c *Component) UpdatePath() {
// 	pathName := []string{}

// 	comp := c
// 	for comp != nil && comp.ComponentSubstationClass != PrimarySubstation {
// 		pathName = append(pathName, comp.ComponentPathname)
// 		comp = comp.Parent
// 	}
// 	if comp != nil {
// 		pathName = append(pathName, c.ComponentPathname)
// 	}

// 	// Reverse the pathName
// 	for i, j := 0, len(pathName)-1; i < j; i, j = i+1, j-1 {
// 		pathName[i], pathName[j] = pathName[j], pathName[i]
// 	}

// 	c.path = strings.Join(pathName, ":")

// }

func (c *Component) GetShortPath() string {
	pathName := []string{}

	comp := c
	for comp != nil && comp.ComponentSubstationClass != PrimarySubstation {
		pathName = append(pathName, comp.ComponentPathname)
		comp = comp.Parent
	}
	if comp != nil {
		pathName = append(pathName, comp.ComponentPathname)
	}

	// Reverse the pathName
	for i, j := 0, len(pathName)-1; i < j; i, j = i+1, j-1 {
		pathName[i], pathName[j] = pathName[j], pathName[i]
	}

	return strings.Join(pathName, ":")
}

func (c *Component) GetFullPath() string {
	pathName := []string{}

	comp := c
	for comp != nil {
		pathName = append(pathName, comp.ComponentPathname)
		comp = comp.Parent
	}

	// Reverse the pathName
	for i, j := 0, len(pathName)-1; i < j; i, j = i+1, j-1 {
		pathName[i], pathName[j] = pathName[j], pathName[i]
	}

	return strings.Join(pathName, ":")
}

func (c *Component) GetPrimaryCircuitComp() *Component { // TODO: see it this can be replaced with IsCircuit ?

	comp := c

	for comp != nil && comp.ComponentSubstationClass.IsPrimaryCircuit() {
		comp = comp.Parent
	}

	return comp
}

// GetGroupingComp returns the grouping component for the component, this is the point under which any circuit or non circuit components will be placed

func (c *Component) GetGroupingComp() *Component {

	comp := c

	// Find the first grouping point
	for comp != nil && !(comp.ComponentSubstationClass == LocationHolder || comp.ComponentClass == 701) { // - this is now enumerating at 7 in the file.. but the src file is als0 changing it back again ... added both conditions
		comp = comp.Parent
	}

	if comp == nil { //No grouping point found, try for the primary sub
		comp = c

		for comp != nil && comp.ComponentSubstationClass != PrimarySubstation {
			comp = comp.Parent
		}

	}

	return comp
}

func (c *Component) IsRoot() bool {
	return c.Parent == nil
}

func (c *Component) IsSwitch() bool {
	return strings.HasSuffix(c.ComponentAlias, "SWDD")
}

func (c *Component) IsSGT() bool {
	return strings.Contains(c.ComponentAlias, "/SGT")
}

func (c *Component) IsDCB() bool {
	// a DCB will have the string "/DCB/" in the alias
	return strings.Contains(c.ComponentAlias, "/DCB/")
}

func (c *Component) IsSCE1() bool {
	return strings.Contains(c.ComponentAlias, "/SC1E")
}

func (c *Component) IsSCE2() bool {
	return strings.Contains(c.ComponentAlias, "/SC2E")
}

func (c *Component) IsCircuit() bool {
	return c.ComponentSubstationClass == PrimaryCircuitID
}

func (c *Component) Is25kvCB() bool {
	return strings.Contains(c.ComponentAlias, "/025_CB/")
}

func (c *Component) IsLeafNode() bool {
	return len(c.Children) == 0
}
