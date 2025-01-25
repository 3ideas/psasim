package compdb

import (
	"fmt"
	"log/slog"
)

type RollbackAction string

const (
	RenameComponentAction RollbackAction = "RenameComponent"
	MoveComponentAction   RollbackAction = "MoveComponent"
	UpdateAttributeAction RollbackAction = "UpdateAttribute"
	CreateAttributeAction RollbackAction = "CreateAttribute"
	CreateComponentAction RollbackAction = "CreateComponent"
)

type RollbackOperation struct {
	Action   RollbackAction
	Alias    string
	OldState interface{} // Store the old state of the component or attribute
}

func (n *ComponentDb) Rollback() error {
	if len(n.rollbackStack) == 0 {
		return fmt.Errorf("no operations to rollback")
	}

	lastOp := n.rollbackStack[len(n.rollbackStack)-1]
	n.rollbackStack = n.rollbackStack[:len(n.rollbackStack)-1]

	switch lastOp.Action {
	case RenameComponentAction:
		comp, err := n.GetComponent(lastOp.Alias)
		if err != nil {
			slog.Error("Rollback: Rename. Failed to get component", "alias", lastOp.Alias, "error", err)
			return fmt.Errorf("Rollback: Rename. Failed to get component %s: %w", lastOp.Alias, err)
		}
		slog.Info("Rollback: Rename. Restoring old name", "alias", lastOp.Alias, "oldName", lastOp.OldState.(string))
		comp.ComponentPathname = lastOp.OldState.(string)
	case MoveComponentAction:
		comp, err := n.GetComponent(lastOp.Alias)
		if err != nil {
			slog.Error("Rollback: Move. Failed to get component", "alias", lastOp.Alias, "error", err)
			return fmt.Errorf("Rollback: Move. Failed to get component %s: %w", lastOp.Alias, err)
		}
		slog.Info("Rollback: Move. Restoring parent ID", "alias", lastOp.Alias, "oldParentID", lastOp.OldState.(string))
		comp.ComponentParentID = lastOp.OldState.(string)
	case UpdateAttributeAction:
		attrNameValue := lastOp.OldState.(AttributeNameValue)
		slog.Info("Rollback: UpdateAttribute. Updating attribute", "alias", lastOp.Alias, "attrName", attrNameValue.Name, "attrValue", attrNameValue.Value)
		comp, err := n.GetComponent(lastOp.Alias)
		if err != nil {
			slog.Error("Rollback: UpdateAttribute. Failed to get component", "alias", lastOp.Alias, "error", err)
			return fmt.Errorf("Rollback: UpdateAttribute. Failed to get component %s: %w", lastOp.Alias, err)
		}
		attr, err := n.GetComponentAttribute(comp.ComponentID, attrNameValue.Name)
		if err != nil {
			slog.Error("Rollback: UpdateAttribute. Failed to get attribute", "alias", lastOp.Alias, "attrName", attrNameValue.Name, "error", err)
			return fmt.Errorf("Rollback: UpdateAttribute. Failed to get attribute %s: %w", attrNameValue.Name, err)
		}
		attr.AttributeValue = attrNameValue.Value
	case CreateAttributeAction:
		newAttr := lastOp.OldState.(*Attribute)
		// Logic to remove the created attribute
		attrRefID := AttributeID{
			ComponentID:   newAttr.ComponentID,
			AttributeName: newAttr.AttributeName,
		}
		comp, err := n.GetComponent(lastOp.Alias)
		if err != nil {
			slog.Error("Rollback: CreateAttribute. Failed to get component", "alias", lastOp.Alias, "error", err)
			return fmt.Errorf("Rollback: CreateAttribute. Failed to get component %s: %w", lastOp.Alias, err)
		}
		slog.Info("Rollback: CreateAttribute. Removing attribute", "alias", lastOp.Alias, "attrName", newAttr.AttributeName, "compID", comp.ComponentID)
		delete(n.Attributes.attr, attrRefID)
	case CreateComponentAction:
		newComp := lastOp.OldState.(*Component)
		slog.Info("Rollback: CreateNewComp. Removing component", "alias", lastOp.Alias, "ID", newComp.ComponentID)
		err := n.Components.RemoveComponent(newComp.ComponentID)
		if err != nil {
			slog.Error("Rollback: CreateNewComp. Failed to remove component", "alias", lastOp.Alias, "error", err)
			return fmt.Errorf("Rollback: CreateNewComp. Failed to remove component %s: %w", lastOp.Alias, err)
		}
	}

	return nil
}

// Implement RollbackAll method
func (n *ComponentDb) RollbackAll() error {
	if len(n.rollbackStack) == 0 {
		return fmt.Errorf("no operations to Rollback")
	}

	slog.Info("Namer: RollbackAll", "numberOfChanges", len(n.rollbackStack))
	count := 0
	for len(n.rollbackStack) > 0 {
		count++
		slog.Info("RollbackAll", "count", count)
		if err := n.Rollback(); err != nil {
			return fmt.Errorf("error during rollback: %w", err)
		}
	}

	return nil
}

func (n *ComponentDb) SetRollbackPoint() error {
	slog.Info("Setting rollback point")
	n.rollbackPoint = len(n.rollbackStack)
	return nil
}

func (n *ComponentDb) RollbackToPoint() error {
	slog.Info("Rolling back to point")
	for len(n.rollbackStack) > n.rollbackPoint {
		if err := n.Rollback(); err != nil {
			return fmt.Errorf("error during rollback: %w", err)
		}
	}
	return nil
}

func (n *ComponentDb) GetNumberOfChanges() (int, error) {
	return len(n.rollbackStack), nil
}
