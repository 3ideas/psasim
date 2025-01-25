package compdb

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Attribute struct {
	ComponentID         string `db:"COMPONENT_ID"`
	AttributeID         string `db:"ATTRIBUTE_ID"`
	AttributeName       string `db:"ATTRIBUTE_NAME"`
	AttributeIndex      int    `db:"ATTRIBUTE_INDEX"`
	AttributeDefinition string `db:"ATTRIBUTE_DEFINITION"`
	// AttributeLocation       string `db:"ATTRIBUTE_LOCATION"`
	AttributeValue string `db:"ATTRIBUTE_VALUE"`
	AttributeType  string `db:"ATTRIBUTE_TYPE"`
	// AttributeTableSize      int    `db:"ATTRIBUTE_TABLE_SIZE"`
	// AttributeVectorSize     int    `db:"ATTRIBUTE_VECTOR_SIZE"`
	AttributeDeType   string `db:"ATTRIBUTE_DE_TYPE"`
	AttributeAlarmRef string `db:"ATTRIBUTE_ALARM_REF"`
	// AttributeWriteGroup     string `db:"ATTRIBUTE_WRITE_GROUP"`
	// AttributeReadGroup      string `db:"ATTRIBUTE_READ_GROUP"`
	AttributeStatus string `db:"ATTRIBUTE_STATUS"`
	// AttributeCloneID        string `db:"ATTRIBUTE_CLONE_ID"`
	// AttributeGRTVLogging    bool   `db:"ATTRIBUTE_GRTV_LOGGING"`
	// AttributeRDBMSArchiving bool   `db:"ATTRIBUTE_RDBMS_ARCHIVING"`
	// AttributeEventPriority  int    `db:"ATTRIBUTE_EVENT_PRIORITY"`
	// AttributeLoggingClass   string `db:"ATTRIBUTE_LOGGING_CLASS"`
	// ProtectionLevel         string `db:"PROTECTION_LEVEL"`
	// CEEvalMode          string `db:"CE_EVAL_MODE"`
	AttributeAlarmIndex int `db:"ATTRIBUTE_ALARM_INDEX"`
	// AttributeAlarmFilter    string `db:"ATTRIBUTE_ALARM_FILTER"`
	// Source                  string `db:"SOURCE"`
	// Identity                string `db:"IDENTITY"`
	// LastGoodValue           string `db:"LAST_GOOD_VALUE"`
	// StatisticsProfile       string `db:"STATISTICS_PROFILE"`
	// RTCalcPeriodicity       int    `db:"RT_CALC_PERIODICITY"`
	// ValidationGroup         string `db:"VALIDATION_GROUP"`
	// DynamicFlags            string `db:"DYNAMIC_FLAGS"`
}

type AttributeID struct {
	ComponentID   string
	AttributeName string
}

type AttributeValue struct {
	Name       string
	Value      string
	ID         string
	CompID     string
	Definition string
}

func (a AttributeValue) String() string {
	return fmt.Sprintf("Name: %s  Value: %s  CompID: %s  AttrID: %s", a.Name, a.Value, a.CompID, a.ID)
}

type Attributes struct {
	attr map[AttributeID]*Attribute
}

func NewAttributeManager() *Attributes {
	return &Attributes{
		attr: make(map[AttributeID]*Attribute),
	}
}

func (a *Attributes) AddAttribute(attr *Attribute) {
	a.attr[AttributeID{ComponentID: attr.ComponentID, AttributeName: attr.AttributeName}] = attr
}

func (a *Attributes) GetAttribute(componentID string, attributeName string) (*Attribute, error) {
	attrID := AttributeID{ComponentID: componentID, AttributeName: attributeName}
	attr, ok := a.attr[attrID]
	if !ok {
		return nil, fmt.Errorf("attribute not found")
	}
	return attr, nil
}

func (a *Attributes) DeleteAttribute(componentID string, attributeName string) {
	attrID := AttributeID{ComponentID: componentID, AttributeName: attributeName}
	delete(a.attr, attrID)
}

func (n *ComponentDb) GetComponentAttribute(componentID string, attributeName string) (*Attribute, error) {
	attrID := AttributeID{ComponentID: componentID, AttributeName: attributeName}
	attr, ok := n.Attributes.attr[attrID]
	if !ok {
		return nil, fmt.Errorf("attribute not found")
	}
	return attr, nil
	// var compAttr Attribute
	// err := n.db.Get(&compAttr, `
	// 	SELECT
	// 		COALESCE(COMPONENT_ID, '') AS COMPONENT_ID,
	// 		COALESCE(ATTRIBUTE_ID, '') AS ATTRIBUTE_ID,
	// 		COALESCE(ATTRIBUTE_NAME, '') AS ATTRIBUTE_NAME,
	// 		COALESCE(ATTRIBUTE_INDEX, 0) AS ATTRIBUTE_INDEX,
	// 		COALESCE(ATTRIBUTE_DEFINITION, '') AS ATTRIBUTE_DEFINITION,
	// 		COALESCE(ATTRIBUTE_LOCATION, '') AS ATTRIBUTE_LOCATION,
	// 		COALESCE(ATTRIBUTE_VALUE, '') AS ATTRIBUTE_VALUE,
	// 		COALESCE(ATTRIBUTE_TYPE, '') AS ATTRIBUTE_TYPE,
	// 		COALESCE(ATTRIBUTE_TABLE_SIZE, 0) AS ATTRIBUTE_TABLE_SIZE,
	// 		COALESCE(ATTRIBUTE_VECTOR_SIZE, 0) AS ATTRIBUTE_VECTOR_SIZE,
	// 		COALESCE(ATTRIBUTE_DE_TYPE, '') AS ATTRIBUTE_DE_TYPE,
	// 		COALESCE(ATTRIBUTE_ALARM_REF, '') AS ATTRIBUTE_ALARM_REF,
	// 		COALESCE(ATTRIBUTE_WRITE_GROUP, '') AS ATTRIBUTE_WRITE_GROUP,
	// 		COALESCE(ATTRIBUTE_READ_GROUP, '') AS ATTRIBUTE_READ_GROUP,
	// 		COALESCE(ATTRIBUTE_STATUS, '') AS ATTRIBUTE_STATUS,
	// 		COALESCE(ATTRIBUTE_CLONE_ID, '') AS ATTRIBUTE_CLONE_ID,
	// 		COALESCE(CASE ATTRIBUTE_GRTV_LOGGING WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS ATTRIBUTE_GRTV_LOGGING,
	// 		COALESCE(CASE ATTRIBUTE_RDBMS_ARCHIVING WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS ATTRIBUTE_RDBMS_ARCHIVING,
	// 		COALESCE(ATTRIBUTE_EVENT_PRIORITY, 0) AS ATTRIBUTE_EVENT_PRIORITY,
	// 		COALESCE(ATTRIBUTE_LOGGING_CLASS, '') AS ATTRIBUTE_LOGGING_CLASS,
	// 		COALESCE(PROTECTION_LEVEL, '') AS PROTECTION_LEVEL,
	// 		COALESCE(CE_EVAL_MODE, '') AS CE_EVAL_MODE,
	// 		COALESCE(ATTRIBUTE_ALARM_INDEX, 0) AS ATTRIBUTE_ALARM_INDEX,
	// 		COALESCE(ATTRIBUTE_ALARM_FILTER, '') AS ATTRIBUTE_ALARM_FILTER,
	// 		COALESCE(SOURCE, '') AS SOURCE,
	// 		COALESCE(IDENTITY, '') AS IDENTITY,
	// 		COALESCE(LAST_GOOD_VALUE, '') AS LAST_GOOD_VALUE,
	// 		COALESCE(STATISTICS_PROFILE, '') AS STATISTICS_PROFILE,
	// 		COALESCE(RT_CALC_PERIODICITY, 0) AS RT_CALC_PERIODICITY,
	// 		COALESCE(VALIDATION_GROUP, '') AS VALIDATION_GROUP,
	// 		COALESCE(DYNAMIC_FLAGS, '') AS DYNAMIC_FLAGS
	// 	FROM COMPONENT_ATTRIBUTES
	// 	WHERE COMPONENT_ID = ? AND ATTRIBUTE_NAME = ?`, componentID, attributeName)
	// if err != nil {
	// 	return nil, err
	// }

	// return &compAttr, nil
}

func GetAttributes(db *sqlx.DB, attributeNames []string) (*Attributes, error) {

	query, args, err := sqlx.In(
		// `SELECT
		// 	COALESCE(COMPONENT_ID, '') AS COMPONENT_ID,
		// 	COALESCE(ATTRIBUTE_NAME, '') AS ATTRIBUTE_NAME,
		// 	COALESCE(ATTRIBUTE_ID, '') AS ATTRIBUTE_ID,
		// 	COALESCE(ATTRIBUTE_INDEX, 0) AS ATTRIBUTE_INDEX,
		// 	COALESCE(ATTRIBUTE_DEFINITION, '') AS ATTRIBUTE_DEFINITION,
		// 	COALESCE(ATTRIBUTE_LOCATION, '') AS ATTRIBUTE_LOCATION,
		// 	COALESCE(ATTRIBUTE_VALUE, '') AS ATTRIBUTE_VALUE,
		// 	COALESCE(ATTRIBUTE_TYPE, '') AS ATTRIBUTE_TYPE,
		// 	COALESCE(ATTRIBUTE_TABLE_SIZE, 0) AS ATTRIBUTE_TABLE_SIZE,
		// 	COALESCE(ATTRIBUTE_VECTOR_SIZE, 0) AS ATTRIBUTE_VECTOR_SIZE,
		// 	COALESCE(ATTRIBUTE_DE_TYPE, '') AS ATTRIBUTE_DE_TYPE,
		// 	COALESCE(ATTRIBUTE_ALARM_REF, '') AS ATTRIBUTE_ALARM_REF,
		// 	COALESCE(ATTRIBUTE_WRITE_GROUP, '') AS ATTRIBUTE_WRITE_GROUP,
		// 	COALESCE(ATTRIBUTE_READ_GROUP, '') AS ATTRIBUTE_READ_GROUP,
		// 	COALESCE(ATTRIBUTE_STATUS, '') AS ATTRIBUTE_STATUS,
		// 	COALESCE(ATTRIBUTE_CLONE_ID, '') AS ATTRIBUTE_CLONE_ID,
		// 	COALESCE(CASE ATTRIBUTE_GRTV_LOGGING WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS ATTRIBUTE_GRTV_LOGGING,
		// 	COALESCE(CASE ATTRIBUTE_RDBMS_ARCHIVING WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS ATTRIBUTE_RDBMS_ARCHIVING,
		// 	COALESCE(ATTRIBUTE_EVENT_PRIORITY, 0) AS ATTRIBUTE_EVENT_PRIORITY,
		// 	COALESCE(ATTRIBUTE_LOGGING_CLASS, '') AS ATTRIBUTE_LOGGING_CLASS,
		// 	COALESCE(PROTECTION_LEVEL, '') AS PROTECTION_LEVEL,
		// 	COALESCE(CE_EVAL_MODE, '') AS CE_EVAL_MODE,
		// 	COALESCE(ATTRIBUTE_ALARM_INDEX, 0) AS ATTRIBUTE_ALARM_INDEX,
		// 	COALESCE(ATTRIBUTE_ALARM_FILTER, '') AS ATTRIBUTE_ALARM_FILTER,
		// 	COALESCE(SOURCE, '') AS SOURCE,
		// 	COALESCE(IDENTITY, '') AS IDENTITY,
		// 	COALESCE(LAST_GOOD_VALUE, '') AS LAST_GOOD_VALUE,
		// 	COALESCE(STATISTICS_PROFILE, '') AS STATISTICS_PROFILE,
		// 	COALESCE(RT_CALC_PERIODICITY, 0) AS RT_CALC_PERIODICITY,
		// 	COALESCE(VALIDATION_GROUP, '') AS VALIDATION_GROUP,
		// 	COALESCE(DYNAMIC_FLAGS, '') AS DYNAMIC_FLAGS
		// FROM COMPONENT_ATTRIBUTES
		// WHERE ATTRIBUTE_NAME IN (?)`,
		`SELECT 
			COALESCE(COMPONENT_ID, '') AS COMPONENT_ID,
			COALESCE(ATTRIBUTE_NAME, '') AS ATTRIBUTE_NAME,
			COALESCE(ATTRIBUTE_ID, '') AS ATTRIBUTE_ID,
			COALESCE(ATTRIBUTE_INDEX, 0) AS ATTRIBUTE_INDEX,
			COALESCE(ATTRIBUTE_VALUE, '') AS ATTRIBUTE_VALUE,
			COALESCE(ATTRIBUTE_TYPE, '') AS ATTRIBUTE_TYPE,
			COALESCE(ATTRIBUTE_DE_TYPE, '') AS ATTRIBUTE_DE_TYPE,
			COALESCE(ATTRIBUTE_ALARM_REF, '') AS ATTRIBUTE_ALARM_REF,
			COALESCE(ATTRIBUTE_STATUS, '') AS ATTRIBUTE_STATUS,
			COALESCE(ATTRIBUTE_ALARM_INDEX, 0) AS ATTRIBUTE_ALARM_INDEX,
			COALESCE(ATTRIBUTE_DEFINITION, '') AS ATTRIBUTE_DEFINITION
		FROM COMPONENT_ATTRIBUTES 
		WHERE ATTRIBUTE_NAME IN (?)`,
		attributeNames)
	if err != nil {
		return nil, err
	}

	var attributes []*Attribute
	err = db.Select(&attributes, query, args...)
	if err != nil {
		return nil, err
	}

	attrManager := NewAttributeManager()

	for _, attr := range attributes {
		attrManager.AddAttribute(attr)

	}

	return attrManager, nil
}
