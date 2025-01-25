package compdb

import (
	"fmt"
)

type ComponentInfo struct {
	Alias               string
	Path                string
	ComponentClassName  string
	SubstationClassName string
	ID                  string
	CloneID             string
	ClonePathname       string
	NameRule            string
	InSymbol            bool
}

func (c *ComponentInfo) String() string {
	return fmt.Sprintf("Alias: %-35s, Path: %-20s Class: %-30s, SClass: %-30s ID: %-30s CloneID: %-14s  ClonePathname: %-20s NameRule: %-20s InSymbol: %t", c.Alias, c.Path, c.ComponentClassName, c.SubstationClassName, c.ID, c.CloneID, c.ClonePathname, c.NameRule, c.InSymbol)
}

type Hierarchy []*ComponentInfo

type NameWithHierachy struct {
	NameDetails
	Hierarchy // First is the component itself, rest are parents all the way to ROOT
}

func (n *ComponentDb) GetNameWithHierarchy(alias string) (*NameWithHierachy, error) {

	nameFull, err := n.GetNameFull(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting name for component %s, %w", alias, err)
	}

	ParentsWithInfo := make([]*ComponentInfo, len(nameFull.Parents))

	for i, comp := range nameFull.Parents {
		compInfo, err := n.GetComponentInfoByID(comp.ComponentID)
		if err != nil {
			return nil, fmt.Errorf("error getting component by ID %s, %w", comp.ComponentID, err)
		}
		ParentsWithInfo[i] = compInfo
	}

	return &NameWithHierachy{NameDetails: nameFull.NameDetails, Hierarchy: ParentsWithInfo}, nil

}

func (n *ComponentDb) GetHierarchyByAlias(alias string) (Hierarchy, error) {

	// comp, err := n.GetComponent(alias)
	// if err != nil {
	// 	return nil, fmt.Errorf("error getting component by ID %s, %w", id, err)
	// }
	// alias := comp.ComponentAlias
	parents, err := n.GetParents(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting parents for component %s, %w", alias, err)
	}

	compInfos := make(Hierarchy, len(parents))

	for i, parent := range parents {
		compInfo, err := n.GetComponentInfoByID(parent.ComponentID)
		if err != nil {
			return nil, fmt.Errorf("error getting component info by ID %s, %w", parent.ComponentID, err)
		}
		compInfos[i] = compInfo
	}
	return compInfos, nil
}

func (n *ComponentDb) GetComponentInfoByID(id string) (*ComponentInfo, error) {

	comp, err := n.GetComponentByID(id)
	if err != nil {
		return nil, fmt.Errorf("error getting component by ID %s, %w", id, err)
	}

	clone, err := n.GetComponentByID(comp.ComponentCloneID)

	clonePathname := ""

	if err == nil && clone != nil {
		clonePathname = clone.ComponentPathname
	}

	classInfo := n.classDefByIndex[comp.ComponentClass]
	componentClassName := classInfo.ComponentClassName

	inSymbol := n.IsInSymbol(comp)

	compInfo := ComponentInfo{
		Alias:               comp.ComponentAlias,
		Path:                comp.ComponentPathname,
		ComponentClassName:  componentClassName,
		SubstationClassName: comp.ComponentSubstationClass.String(),
		ID:                  comp.ComponentID,
		CloneID:             comp.ComponentCloneID,
		ClonePathname:       clonePathname,
		NameRule:            classInfo.ComponentNameRule,
		InSymbol:            inSymbol,
	}

	return &compInfo, nil
}

func (n *ComponentDb) GetComponentInfo(alias string) (*ComponentInfo, error) {

	comp, err := n.GetComponent(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting component by alias %s, %w", alias, err)
	}

	clone, err := n.GetComponentByID(comp.ComponentCloneID)

	clonePathname := ""

	if err == nil && clone != nil {
		clonePathname = clone.ComponentPathname
	}

	classInfo := n.classDefByIndex[comp.ComponentClass]
	componentClassName := classInfo.ComponentClassName

	inSymbol := n.IsInSymbol(comp)

	compInfo := ComponentInfo{
		Alias:               comp.ComponentAlias,
		Path:                comp.ComponentPathname,
		ComponentClassName:  componentClassName,
		SubstationClassName: comp.ComponentSubstationClass.String(),
		ID:                  comp.ComponentID,
		CloneID:             comp.ComponentCloneID,
		ClonePathname:       clonePathname,
		NameRule:            classInfo.ComponentNameRule,
		InSymbol:            inSymbol,
	}

	return &compInfo, nil
}

// IsInSymbol Determin if a component is in a multi comp symbol
// it does this by checking if the parent of the component clone is the same as the components clone's parent
// or if the the component has children then if a child and parent clones ....
func (n *ComponentDb) IsInSymbol(comp *Component) bool {

	if comp.ComponentAlias == "ROOT" { // Probably not required but does no harm.
		return false
	}

	clone, err := n.GetComponentByID(comp.ComponentCloneID)
	if err != nil {
		return false //  Not all components have clones!
	}

	// Two ways to determin if a component is in a multi comp symbol
	// if its parent has a clone  and it matches the components clone parent
	parent, err := n.GetComponentByID(comp.ComponentParentID)
	if err == nil {
		// If parent clone ID matches the components clones parent then this has been cloned from a symbol
		if clone.ComponentParentID == parent.ComponentCloneID {
			return true
		}
	}

	// or if the component has children then if a child and parent clones ...
	children := comp.Children
	if len(children) == 0 {
		return false // No children so not in symbol
	}

	// Only need to check one child
	child := children[0]
	childClone, err := n.GetComponentByID(child.ComponentCloneID)
	if err != nil {
		return false
	}

	if childClone.ComponentParentID == clone.ComponentID {
		return true
	}

	return false
}

func (nh *NameWithHierachy) String() string {
	return nh.NameDetails.String()
}

func (h Hierarchy) InSymbol() bool {
	return h[0].InSymbol
}

func (h Hierarchy) IsCompToMoveDifferentFromCurrentComp() bool {
	s := h.GetSymbolComponent()

	return s.Alias != h[0].Alias
}

func (h Hierarchy) GetSymbolComponent() *ComponentInfo {
	if !h.InSymbol() {
		return h[0]
	}
	lastComp := h[0]
	for _, comp := range h {
		if !comp.InSymbol {
			return lastComp
		}
		lastComp = comp
	}
	return nil
}

func (n *ComponentDb) GetChildrenInfoByID(id string) ([]*ComponentInfo, error) {

	comp, err := n.GetComponentByID(id)
	if err != nil {
		return nil, fmt.Errorf("error getting component by ID %s, %w", id, err)
	}
	children := comp.Children
	if len(children) == 0 {
		return nil, fmt.Errorf("no children found for ID %s", id)
	}

	childrenInfo := make([]*ComponentInfo, len(children))
	for i, child := range children {
		childInfo, err := n.GetComponentInfoByID(child.ComponentID)
		if err != nil {
			return nil, err
		}
		childrenInfo[i] = childInfo
	}
	return childrenInfo, nil
}
