package compdb

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TesrRollback(t *testing.T) {
	localNamer := NewCompDb()

	// Setup: Create a component
	alias := "comp1"
	rootComp := Component{
		ComponentID:       uuid.New().String(),
		ComponentPathname: "OldName",
		ComponentAlias:    alias,
	}
	localNamer.Components.AddComponent(&rootComp)

	// Test Rename
	err := localNamer.RenameComponent(alias, "NewName")
	assert.NoError(t, err)
	assert.Equal(t, "NewName", rootComp.ComponentPathname)

	// Test Rollback
	err = localNamer.Rollback()
	assert.NoError(t, err)
	assert.Equal(t, "OldName", rootComp.ComponentPathname)

	// Test CreateAttribute
	attrName := "attr1"
	attrValue := "value1"
	err = localNamer.CreateAttribute(alias, attrName, attrValue)
	assert.NoError(t, err)

	// Test Rollback for CreateAttribute
	err = localNamer.Rollback()
	assert.NoError(t, err)
	_, err = localNamer.GetComponentAttribute(rootComp.ComponentID, attrName)
	assert.Error(t, err) // Attribute should be removed

	// Test Move
	// Add a new component as a child of comp1
	newComp2Alias := "comp2"
	newComp2 := Component{
		ComponentID:       uuid.New().String(),
		ComponentPathname: "NewComp2Name",
		ComponentAlias:    newComp2Alias,
		ComponentParentID: rootComp.ComponentID,
	}
	localNamer.Components.AddComponent(&newComp2)
	newComp3Alias := "comp3"
	newComp3 := Component{
		ComponentID:       uuid.New().String(),
		ComponentPathname: "NewComp3Name",
		ComponentAlias:    newComp3Alias,
		ComponentParentID: rootComp.ComponentID,
	}
	// Add a child to the new component
	newCompChildAlias := "comp4"
	newCompChild := Component{
		ComponentID:       uuid.New().String(),
		ComponentPathname: "NewCompChildName",
		ComponentAlias:    newCompChildAlias,
		ComponentParentID: newComp3.ComponentID,
	}
	localNamer.Components.AddComponent(&newCompChild)

	err = localNamer.MoveComponent(newCompChildAlias, newComp2Alias)
	assert.NoError(t, err)
	assert.Equal(t, newCompChild.ComponentParentID, newComp2.ComponentID)

	// Test Rollback for Move
	err = localNamer.Rollback()
	assert.NoError(t, err)
	assert.Equal(t, newCompChild.ComponentParentID, newComp3.ComponentID) // Assuming the original parent ID is empty

	// Test CreateNewComp
	err = localNamer.CreateComponent("newComp", "NewCompName", alias, "", "")
	assert.NoError(t, err)

	// Test Rollback for CreateNewComp
	err = localNamer.Rollback()
	assert.NoError(t, err)
	// Check if the new component is removed (you may need to implement a way to check this)
}

func TestRollbackAll(t *testing.T) {
	localNamer := NewCompDb()

	// Setup: Create a component
	alias := "comp1"
	comp := Component{
		ComponentID:       uuid.New().String(),
		ComponentPathname: "OldName",
		ComponentAlias:    alias,
	}
	localNamer.Components.AddComponent(&comp)

	// Perform multiple operations
	localNamer.RenameComponent(alias, "NewName")
	localNamer.CreateAttribute(alias, "attr1", "value1")
	localNamer.MoveComponent(alias, "newParent")
	localNamer.CreateComponent("newComp", "NewCompName", alias, "", "")

	// Test RollbackAll
	err := localNamer.RollbackAll()
	assert.NoError(t, err)

	// Check if the component is back to its original state
	assert.Equal(t, "OldName", comp.ComponentPathname)
	_, err = localNamer.GetComponentAttribute(comp.ComponentID, "attr1")
	assert.Error(t, err)                        // Attribute should be removed
	assert.Equal(t, "", comp.ComponentParentID) // Assuming the original parent ID is empty
	// Check if the new component is removed (you may need to implement a way to check this)
}

func TestSetRollbackPoint(t *testing.T) {
	localNamer := NewCompDb()

	localNamer.SetRollbackPoint()
	// Setup: Create a component
	localNamer.CreateComponent("ROOT", "ROOT", "", "", "")

	localNamer.RollbackToPoint()

	noOfChanges, err := localNamer.GetNumberOfChanges()
	if err != nil {
		t.Errorf("Error getting number of changes: %v", err)
	}
	assert.Equal(t, 0, noOfChanges)

	localNamer.CreateComponent("ROOT", "ROOT", "", "", "")
	localNamer.CreateComponent("comp1", "comp1", "ROOT", "", "")

	noOfChanges, err = localNamer.GetNumberOfChanges()
	if err != nil {
		t.Errorf("Error getting number of changes: %v", err)
	}
	assert.Equal(t, 2, noOfChanges)

	localNamer.SetRollbackPoint()

	localNamer.CreateComponent("comp2", "comp2", "comp1", "", "")

	noOfChanges, err = localNamer.GetNumberOfChanges()
	if err != nil {
		t.Errorf("Error getting number of changes: %v", err)
	}
	assert.Equal(t, 3, noOfChanges)

	localNamer.RollbackToPoint()

	noOfChanges, err = localNamer.GetNumberOfChanges()
	if err != nil {
		t.Errorf("Error getting number of changes: %v", err)
	}
	assert.Equal(t, 2, noOfChanges)

}

func TestRollbackToPoint(t *testing.T) {
	localNamer := NewCompDb()

	root := "ROOT"
	localNamer.CreateComponent(root, root, "", "", "")

	comp1 := "comp1"
	localNamer.CreateComponent(comp1, comp1, root, "", "")
	comp2 := "comp2"
	localNamer.CreateComponent(comp2, comp2, root, "", "")

	// Perform some operations

	localNamer.RenameComponent(comp1, "NewName")
	localNamer.CreateAttribute(comp1, "attr1", "value1")

	noOfChanges, _ := localNamer.GetNumberOfChanges()
	assert.Equal(t, 5, noOfChanges) // 4 operations performed

	// Set rollback point
	err := localNamer.SetRollbackPoint()
	assert.NoError(t, err)

	// Perform more operations
	localNamer.MoveComponent(comp1, comp2)

	// Check the number of changes before RollbackToPoint
	noOfChanges, _ = localNamer.GetNumberOfChanges()
	assert.Equal(t, 6, noOfChanges) // 4 operations performed

	// Rollback to the set point
	err = localNamer.RollbackToPoint()
	assert.NoError(t, err)

	// Check the state of the component after rollback
	comp, err := localNamer.GetComponent(comp1)
	assert.NoError(t, err)
	rootComp, err := localNamer.GetComponent(root)

	assert.NoError(t, err)
	assert.Equal(t, rootComp.ComponentID, comp.ComponentParentID) // Should revert to parent
	assert.Equal(t, "NewName", comp.ComponentPathname)            // should still be unchanged

}
