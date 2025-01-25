package comps

import (
	"fmt"
)

func (c *ComponentManager) AddRollbackEntry(rollbackOperation RollbackOperation) {
	c.rollbackStack = append(c.rollbackStack, rollbackOperation)
}

func (c *ComponentManager) CreateComponent(alias string, name string, path string, parent *Component, substationClass string) (*Component, error) {
	newComp := &Component{Path: path, Alias: alias, Name: name, SubstationClass: substationClass, ComponentClass: substationClass, Parent: parent} // Not sure if setting parent is needed here
	// lastCompSelected = newComp
	err := c.Add(newComp)
	if err != nil {
		return nil, fmt.Errorf("error adding component: %s", err)
	}

	// Push operation to rollback stack
	c.rollbackStack = append(c.rollbackStack, RollbackOperation{CreateComponentAction, alias, nil}) // TODO: change name of operation to CreateComponent

	return newComp, nil
}

func (c *ComponentManager) RenameComponent(alias string, newName string) error {
	comp, ok := c.GetCompByAlias(alias)
	if !ok {
		return fmt.Errorf("no such component: %s", alias)
	}

	originalName := comp.Name

	// Remove from c.Components.pathToComp
	delete(c.pathToComp, comp.Path)

	comp.Name = newName
	// Update path
	comp.Path = getParentPath(comp.Path) + ":" + newName
	// update c.Components.pathToComp with the new path
	c.pathToComp[comp.Path] = comp

	// Push operation to rollback stack
	c.rollbackStack = append(c.rollbackStack, RollbackOperation{RenameComponentAction, alias, originalName})
	return nil
}

func (c *ComponentManager) MoveComponent(alias, newParentAlias string) error {
	comp, ok := c.GetCompByAlias(alias)
	if !ok {
		return fmt.Errorf("no such component: %s", alias)
	}

	newParent, ok := c.GetCompByAlias(newParentAlias)
	if !ok {
		return fmt.Errorf("no such component: %s", newParentAlias)
	}

	oldParentAlias := comp.Parent.Alias
	originalParent := comp.Parent
	comp.OriginalParent = originalParent

	// Remove from the old parent
	originalParent.RemoveChild(comp)
	// Set the new parent
	newParent.AddChild(comp)

	// Push operation to rollback stack
	c.rollbackStack = append(c.rollbackStack, RollbackOperation{MoveComponentAction, alias, oldParentAlias})
	return nil
}
