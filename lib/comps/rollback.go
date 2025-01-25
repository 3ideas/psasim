package comps

import (
	"fmt"
	"log/slog"
)

type RollbackAction string

const (
	RenameComponentAction RollbackAction = "RenameComponent"
	MoveComponentAction   RollbackAction = "MoveComponent"
	CreateComponentAction RollbackAction = "CreateComponent"
)

type RollbackOperation struct {
	Action   RollbackAction
	Alias    string
	OldState interface{} // Store the old state of the component or attribute
}

func (c *ComponentManager) Rollback() error {
	if len(c.rollbackStack) == 0 {
		return fmt.Errorf("ComponentManager Rollback: no operations to rollback")
	}

	lastOp := c.rollbackStack[len(c.rollbackStack)-1]
	c.rollbackStack = c.rollbackStack[:len(c.rollbackStack)-1]

	switch lastOp.Action {
	case RenameComponentAction:
		comp, ok := c.GetCompByAlias(lastOp.Alias)
		if !ok {
			slog.Error("ComponentManager Rollback: Rename. Failed to get component", "alias", lastOp.Alias)
			return fmt.Errorf("ComponentManager Rollback: Rename. Failed to get component %s", lastOp.Alias)
		}
		slog.Info("Rollback: Rename. Restoring old name", "alias", lastOp.Alias, "oldName", lastOp.OldState.(string))

		// Remove from c.Components.pathToComp
		delete(c.pathToComp, comp.Path)

		comp.Name = lastOp.OldState.(string)
		comp.Path = getParentPath(comp.Path) + ":" + comp.Name
		// update c.Components.pathToComp with the new path
		c.pathToComp[comp.Path] = comp

	case MoveComponentAction:
		comp, ok := c.GetCompByAlias(lastOp.Alias)
		if !ok {
			slog.Error("ComponentManager Rollback: Move. Failed to get component", "alias", lastOp.Alias)
			return fmt.Errorf("Rollback: Move. Failed to get component %s", lastOp.Alias)
		}

		originalParentAlias := lastOp.OldState.(string)

		// Get the parnet to move to
		newParent, ok := c.GetCompByAlias(originalParentAlias)
		if !ok {
			slog.Error("ComponentManager Rollback: Move. Failed to get new parent", "alias", originalParentAlias)
			return fmt.Errorf("ComponentManager Rollback: Move. Failed to get new parent %s", originalParentAlias)
		}

		slog.Info("ComponentManager Rollback: Move. Restoring parent ID", "alias", lastOp.Alias, "oldParentID", lastOp.OldState.(string))

		// oldParentAlias := comp.Parent.Alias
		originalParent := comp.Parent
		comp.OriginalParent = nil // Reset the original parent this was used to track to see if comp had been moved TODO: remove this?

		// Remove from the old parent
		originalParent.RemoveChild(comp)
		// Set the new parent
		newParent.AddChild(comp)

	case CreateComponentAction:
		alias := lastOp.Alias
		slog.Info("Rollback: CreateNewComp. Removing component", "alias", alias)
		err := c.RemoveComponent(alias)
		if err != nil {
			slog.Error("Rollback: CreateNewComp. Failed to remove component", "alias", alias, "error", err)
			return fmt.Errorf("Rollback: CreateNewComp. Failed to remove component %s: %w", lastOp.Alias, err)
		}
	}

	return nil
}

func (c *ComponentManager) SetRollbackPoint() error {
	slog.Debug("ComponentManager: Setting rollback point")
	c.rollbackPoint = len(c.rollbackStack)
	return nil
}

func (c *ComponentManager) RollbackToPoint() error {
	slog.Debug("ComponentManager:Rolling back to point")
	for len(c.rollbackStack) > c.rollbackPoint {
		if err := c.Rollback(); err != nil {
			return fmt.Errorf("error during rollback: %w", err)
		}
	}
	return nil
}
