package namerif

import "github.com/3ideas/psasim/lib/compdb"

type NameService interface {
	GetName(alias string) (*compdb.NameDetails, error)
	GetNameWithHierarchy(alias string) (*compdb.NameWithHierachy, error)
	RenameComponent(alias, newName string) error
	MoveComponent(alias, newLocationAlias string) error
	CreateAttribute(alias, attrName, attrValue string) error
	UpdateAttribute(alias, attrName, attrValue string) error
	CreateComponent(alias, name, parentAlias, templateAlias, substationClassName string) error
	RollbackAll() error
	GetNumberOfChanges() (int, error)
	GetAttributeValue(alias, attrName string) (compdb.AttributeValue, error)
	GetComponentClassDetails(alias string) (*compdb.ComponentClassDetails, error)
	SetRollbackPoint() error
	RollbackToPoint() error
	// GetComponentInfoByAlias(alias string) (*namer.ComponentInfo, error)
	GetComponentInfoByID(id string) (*compdb.ComponentInfo, error)
	GetComponentInfo(alias string) (*compdb.ComponentInfo, error)
	GetChildrenInfoByID(id string) ([]*compdb.ComponentInfo, error)

	GetHierarchyByAlias(alias string) (compdb.Hierarchy, error)
}
