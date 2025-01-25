package compdb

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type AttributeNameValue struct {
	Name  string
	Value string
}

func (n *ComponentDb) RenameComponent(alias, newName string) error {
	comp, err := n.GetComponent(alias)
	if err != nil {
		return fmt.Errorf("error getting component %s: %w", alias, err)
	}
	oldName := comp.ComponentPathname
	comp.ComponentPathname = newName

	slog.Info("Namer: Rename", "alias", alias, "oldName", oldName, "newName", newName)

	// Push operation to rollback stack
	n.rollbackStack = append(n.rollbackStack, RollbackOperation{RenameComponentAction, alias, oldName})

	return nil
}

func (n *ComponentDb) MoveComponent(alias, newLocationAlias string) error {
	comp, err := n.GetComponent(alias)
	if err != nil {
		return fmt.Errorf("error getting component %s: %w", alias, err)
	}
	oldParentID := comp.ComponentParentID

	newLocation, err := n.GetComponent(newLocationAlias)
	if err != nil {
		return fmt.Errorf("error getting component %s: %w", newLocationAlias, err)
	}

	slog.Info("Namer: Move", "alias", alias, "newLocationAlias", newLocationAlias)

	comp.ComponentParentID = newLocation.ComponentID

	// Push operation to rollback stack
	n.rollbackStack = append(n.rollbackStack, RollbackOperation{MoveComponentAction, alias, oldParentID})

	return nil
}

func (n *ComponentDb) CreateAttribute(alias, attrName, attrValue string) error {
	comp, err := n.GetComponent(alias)
	if err != nil {
		return fmt.Errorf("error getting component %s: %w", alias, err)
	}
	attr, err := n.GetComponentAttribute(comp.ComponentID, attrName)
	if err == nil {
		if attr.AttributeValue == attrValue {
			slog.Info("CreateAttribute: Attribute already exists, with same value, skipping", "alias", alias, "attrName", attrName, "attrValue", attr.AttributeValue, "NewValue", attrValue)
			return nil
		}
		slog.Info("CreateAttribute: Attribute already exists, updating value", "alias", alias, "attrName", attrName, "attrValue", attr.AttributeValue, "NewValue", attrValue)

		oldValue := attr.AttributeValue
		attr.AttributeValue = attrValue

		// Push operation to rollback stack
		n.rollbackStack = append(n.rollbackStack, RollbackOperation{UpdateAttributeAction, alias, AttributeNameValue{Name: attrName, Value: oldValue}})
		slog.Info("CreateAttribute: Update", "alias", alias, "attrName", attrName, "oldValue", oldValue, "newValue", attrValue)
	} else {
		// Generate uniq ID for the attribute
		attrID := uuid.New().String()
		attr = &Attribute{
			ComponentID:    comp.ComponentID,
			AttributeID:    attrID,
			AttributeName:  attrName,
			AttributeValue: attrValue,
		}
		attrRefID := AttributeID{
			ComponentID:   comp.ComponentID,
			AttributeName: attrName,
		}
		n.Attributes.attr[attrRefID] = attr

		slog.Info("Namer: CreateAttribute", "alias", alias, "attrName", attrName, "newValue", attrValue)
		// Push operation to rollback stack
		n.rollbackStack = append(n.rollbackStack, RollbackOperation{CreateAttributeAction, alias, attr})
	}
	// If attrubute name is "Circuit name" then update the component name, this is how PO pfl behaves (really its probably the name of the attribute is in SYS_CIRCUIT_NAME but its hardcoded here)
	if attrName == "Circuit name" {
		n.RenameComponent(alias, attrValue)
	}
	return nil
}

func (n *ComponentDb) UpdateAttribute(alias, attrName, attrValue string) error {
	comp, err := n.GetComponent(alias)
	if err != nil {
		return fmt.Errorf("error getting component %s: %w", alias, err)
	}
	attr, err := n.GetComponentAttribute(comp.ComponentID, attrName)
	if err != nil {
		return fmt.Errorf("error getting attribute %s for component %s: %w", attrName, alias, err)
	}
	if attr.AttributeValue == attrValue {
		slog.Info("UpdateAttribute: Attribute already exists, with same value, skipping", "alias", alias, "attrName", attrName, "attrValue", attr.AttributeValue, "NewValue", attrValue)
		return nil
	}
	slog.Info("UpdateAttribute: updating value", "alias", alias, "attrName", attrName, "attrValue", attr.AttributeValue, "NewValue", attrValue)

	oldValue := attr.AttributeValue
	attr.AttributeValue = attrValue

	// Push operation to rollback stack
	n.rollbackStack = append(n.rollbackStack, RollbackOperation{UpdateAttributeAction, alias, AttributeNameValue{Name: attrName, Value: oldValue}})
	slog.Info("Namer: CreateAttribute Update", "alias", alias, "attrName", attrName, "oldValue", oldValue, "newValue", attrValue)

	if attrName == "Circuit name" { // If attrubute name is "Circuit name" then update the component name, this is how PO pfl behaves (really its probably the name of the attribute is in SYS_CIRCUIT_NAME but its hardcoded here)
		n.RenameComponent(alias, attrValue)
	}

	return nil
}

func (n *ComponentDb) CreateComponentReturnComponent(alias, name, parentAlias, templateAlias, substationClassName string) (*Component, error) {

	parent, err := n.GetComponent(parentAlias)
	if err != nil && (parentAlias != "" && len(n.componentsByAlias) > 0) {
		return nil, fmt.Errorf("error getting component %s: %w", parentAlias, err)
	}

	comp, err := n.GetComponent(alias)
	if err == nil {
		slog.Info("Component already exists, skipping", "alias", alias, "name", comp.ComponentPathname, "ID", comp.ComponentID, "parentAlias", parentAlias, "templateAlias", templateAlias, "substationClassName", substationClassName)
		return comp, nil
	}

	// Ignore error as it will return NotApplicable if the name is not found
	substationClass, _ := GetSubstationClassFromName(substationClassName) // Ignore the error as it will return NotApplicable if the name is not found

	// Generate a new component ID
	compID := uuid.New().String()

	var newComp Component

	if templateAlias != "" {
		template, err := n.GetComponent(templateAlias)
		if err != nil {
			return nil, fmt.Errorf("error getting template component %s: %w", templateAlias, err)
		}
		newComp = *template
	}

	// Create a new component with the template
	newComp.ComponentID = compID
	newComp.ComponentPathname = name
	newComp.ComponentAlias = alias
	// newComp.ComponentName = name

	// newComp.ComponentClass = template.ComponentClass
	newComp.ComponentSubstationClass = substationClass

	parentId := ""
	if parent != nil {
		parentId = parent.ComponentID
	}
	newComp.ComponentParentID = parentId

	n.Components.AddComponent(&newComp)

	slog.Info("Namer: CreateNewComp", "alias", alias, "name", name, "ID", newComp.ComponentID, "ParentID", newComp.ComponentParentID, "ParentAlias", parentAlias, "templateAlias", templateAlias, "substationClassName", substationClassName)

	// Push operation to rollback stack
	n.rollbackStack = append(n.rollbackStack, RollbackOperation{CreateComponentAction, alias, &newComp}) // TODO: change name of operation to CreateComponent

	return &newComp, nil
}

func (n *ComponentDb) CreateComponent(alias, name, parentAlias, templateAlias, substationClassName string) error {
	_, err := n.CreateComponentReturnComponent(alias, name, parentAlias, templateAlias, substationClassName)
	return err
}
