package compdb

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type ComponentClassIndex int

func (c ComponentClassIndex) String() string {
	if componentClassDefinitions != nil {
		compClass, err := componentClassDefinitions.GetComponentClassDefnByIndex(c)
		if err != nil {
			return fmt.Sprintf("UnknownClass(%d)", c)
		}
		return fmt.Sprintf("%s(%d)", compClass.ComponentClassName, c)
	}
	return fmt.Sprintf("%d", c)
}

type ComponentClassDefn struct {
	ComponentClassIndex       ComponentClassIndex `db:"COMPONENT_CLASS_INDEX"`
	ComponentClassName        string              `db:"COMPONENT_CLASS_NAME"`
	ComponentStatus           string              `db:"COMPONENT_STATUS"`
	ComponentAbbreviation     string              `db:"COMPONENT_ABBREVIATION"`
	ComponentAppearance       string              `db:"COMPONENT_APPEARANCE"`
	ComponentLifeCycle        string              `db:"COMPONENT_LIFE_CYCLE"`
	ComponentEMSClassIndex    int                 `db:"COMPONENT_EMS_CLASS_INDEX"`
	ComponentTraceComponent   string              `db:"COMPONENT_TRACE_COMPONENT"`
	ComponentTraceLine        string              `db:"COMPONENT_TRACE_LINE"`
	ComponentIsJunction       bool                `db:"COMPONENT_IS_JUNCTION"`
	ComponentHasCustomers     bool                `db:"COMPONENT_HAS_CUSTOMERS"`
	ComponentNameRule         string              `db:"COMPONENT_NAME_RULE"`
	ComponentIsSupInfeed      bool                `db:"COMPONENT_IS_SUP_INFEED"`
	ComponentIsFeederEquiv    bool                `db:"COMPONENT_IS_FEEDER_EQUIV"`
	ComponentIsAsset          bool                `db:"COMPONENT_IS_ASSET"`
	ComponentIsLocation       bool                `db:"COMPONENT_IS_LOCATION"`
	ComponentIsTransferAttr   bool                `db:"COMPONENT_IS_TRANSFER_ATTR"`
	ComponentIsTransferAlias  bool                `db:"COMPONENT_IS_TRANSFER_ALIAS"`
	ComponentIsTransferName   bool                `db:"COMPONENT_IS_TRANSFER_NAME"`
	ComponentIsTransferParent bool                `db:"COMPONENT_IS_TRANSFER_PARENT"`
	NonPatchable              bool                `db:"NON_PATCHABLE"`
	ComponentRTClass          string              `db:"COMPONENT_RT_CLASS"`
	ComponentDelZoneSharable  bool                `db:"COMPONENT_DEL_ZONE_SHARABLE"`
	ComponentIsLvRelevant     bool                `db:"COMPONENT_IS_LV_RELEVANT"`
	ComponentIsTelemetered    bool                `db:"COMPONENT_IS_TELEMETERED"`
	ComponentSLDClass         string              `db:"COMPONENT_SLD_CLASS"`
	ApplyToConnCompClasses    string              `db:"APPLY_TO_CONN_COMP_CLASSES"`
	TracedNameRule            string              `db:"TRACED_NAME_RULE"`
	TracedNamingPriority      int                 `db:"TRACED_NAMING_PRIORITY"`
	TracedNamingCompNameRule  string              `db:"TRACED_NAMING_COMP_NAME_RULE"`
	ComponentCategory         string              `db:"COMPONENT_CATEGORY"`
	MenuName                  string              `db:"MENU_NAME"`
	ShowInEELocationMode      bool                `db:"SHOW_IN_EE_LOCATION_MODE"`
	IsolationClass            string              `db:"ISOLATION_CLASS"`
	ComponentIsMaintainable   bool                `db:"COMPONENT_IS_MAINTAINABLE"`
	CheckParallel             bool                `db:"CHECK_PARALLEL"`
	MixedPhaseStatesSymbol    string              `db:"MIXED_PHASE_STATES_SYMBOL"`
	ComponentIsLvSwitch       bool                `db:"COMPONENT_IS_LV_SWITCH"`
	TraceClassLookup          string              `db:"TRACE_CLASS_LOOKUP"`
	TooltipName               string              `db:"TOOLTIP_NAME"`
	IsMultiplePositionSwitch  bool                `db:"IS_MULTIPLE_POSITION_SWITCH"`
	ExcludeFromDelCand        bool                `db:"EXCLUDE_FROM_DEL_CAND"`
	GenerateCIMMRID           string              `db:"GENERATE_CIM_MRID"`
	MeasurementSide           string              `db:"MEASUREMENT_SIDE"`
	AutoIntroduceRule         string              `db:"AUTO_INTRODUCE_RULE"`
	OpExchange                string              `db:"OP_EXCHANGE"`
	AutoAlias                 string              `db:"AUTO_ALIAS"`
	ImportUpdateProtection    bool                `db:"IMPORT_UPDATE_PROTECTION"`
	IsTempSCADA               bool                `db:"IS_TEMP_SCADA"`
	StudyTooltipName          string              `db:"STUDY_TOOLTIP_NAME"`
}

type ComponentClassDetails struct {
	ClassName       string
	NameRule        string
	SubstationClass SubstationType
}

type ComponentClassDefns struct {
	classDefByName  map[string]*ComponentClassDefn
	classDefByIndex map[ComponentClassIndex]*ComponentClassDefn
}

// GetComponentClassDefn gets the component class definition for a given component class index
// func GetComponentClassDefn(db *sqlx.DB, componentClassIndex int) (*ComponentClassDefn, error) {
// 	var componentClassDefns ComponentClassDefn
// 	err := db.Get(&componentClassDefns, `
// 		SELECT
// 			COALESCE(COMPONENT_CLASS_INDEX, 0) AS COMPONENT_CLASS_INDEX,
// 			COALESCE(COMPONENT_CLASS_NAME, '') AS COMPONENT_CLASS_NAME,
// 			COALESCE(COMPONENT_STATUS, '') AS COMPONENT_STATUS,
// 			COALESCE(COMPONENT_ABBREVIATION, '') AS COMPONENT_ABBREVIATION,
// 			COALESCE(COMPONENT_APPEARANCE, '') AS COMPONENT_APPEARANCE,
// 			COALESCE(COMPONENT_LIFE_CYCLE, '') AS COMPONENT_LIFE_CYCLE,
// 			COALESCE(COMPONENT_EMS_CLASS_INDEX, 0) AS COMPONENT_EMS_CLASS_INDEX,
// 			COALESCE(COMPONENT_TRACE_COMPONENT, '') AS COMPONENT_TRACE_COMPONENT,
// 			COALESCE(COMPONENT_TRACE_LINE, '') AS COMPONENT_TRACE_LINE,
// 			COALESCE(CASE COMPONENT_IS_JUNCTION WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_JUNCTION,
// 			COALESCE(CASE COMPONENT_HAS_CUSTOMERS WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_HAS_CUSTOMERS,
// 			COALESCE(COMPONENT_NAME_RULE, '') AS COMPONENT_NAME_RULE,
// 			COALESCE(CASE COMPONENT_IS_SUP_INFEED WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_SUP_INFEED,
// 			COALESCE(CASE COMPONENT_IS_FEEDER_EQUIV WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_FEEDER_EQUIV,
// 			COALESCE(CASE COMPONENT_IS_ASSET WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_ASSET,
// 			COALESCE(CASE COMPONENT_IS_LOCATION WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_LOCATION,
// 			COALESCE(CASE COMPONENT_IS_TRANSFER_ATTR WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_TRANSFER_ATTR,
// 			COALESCE(CASE COMPONENT_IS_TRANSFER_ALIAS WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_TRANSFER_ALIAS,
// 			COALESCE(CASE COMPONENT_IS_TRANSFER_NAME WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_TRANSFER_NAME,
// 			COALESCE(CASE COMPONENT_IS_TRANSFER_PARENT WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_TRANSFER_PARENT,
// 			COALESCE(NON_PATCHABLE, false) AS NON_PATCHABLE,
// 			COALESCE(COMPONENT_RT_CLASS, '') AS COMPONENT_RT_CLASS,
// 			COALESCE(CASE COMPONENT_DEL_ZONE_SHARABLE WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_DEL_ZONE_SHARABLE,
// 			COALESCE(CASE COMPONENT_IS_LV_RELEVANT WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_LV_RELEVANT,
// 			COALESCE(CASE COMPONENT_IS_TELEMETERED WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_TELEMETERED,
// 			COALESCE(COMPONENT_SLD_CLASS, '') AS COMPONENT_SLD_CLASS,
// 			COALESCE(APPLY_TO_CONN_COMP_CLASSES, '') AS APPLY_TO_CONN_COMP_CLASSES,
// 			COALESCE(TRACED_NAME_RULE, '') AS TRACED_NAME_RULE,
// 			COALESCE(TRACED_NAMING_PRIORITY, 0) AS TRACED_NAMING_PRIORITY,
// 			COALESCE(TRACED_NAMING_COMP_NAME_RULE, '') AS TRACED_NAMING_COMP_NAME_RULE,
// 			COALESCE(COMPONENT_CATEGORY, '') AS COMPONENT_CATEGORY,
// 			COALESCE(MENU_NAME, '') AS MENU_NAME,
// 			COALESCE(CASE SHOW_IN_EE_LOCATION_MODE WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS SHOW_IN_EE_LOCATION_MODE,
// 			COALESCE(ISOLATION_CLASS, '') AS ISOLATION_CLASS,
// 			COALESCE(CASE COMPONENT_IS_MAINTAINABLE WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_MAINTAINABLE,
// 			COALESCE(CASE CHECK_PARALLEL WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS CHECK_PARALLEL,
// 			COALESCE(MIXED_PHASE_STATES_SYMBOL, '') AS MIXED_PHASE_STATES_SYMBOL,
// 			COALESCE(CASE COMPONENT_IS_LV_SWITCH WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_LV_SWITCH,
// 			COALESCE(TRACE_CLASS_LOOKUP, '') AS TRACE_CLASS_LOOKUP,
// 			COALESCE(TOOLTIP_NAME, '') AS TOOLTIP_NAME,
// 			COALESCE(CASE IS_MULTIPLE_POSITION_SWITCH WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS IS_MULTIPLE_POSITION_SWITCH,
// 			COALESCE(CASE EXCLUDE_FROM_DEL_CAND WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS EXCLUDE_FROM_DEL_CAND,
// 			COALESCE(GENERATE_CIM_MRID, '') AS GENERATE_CIM_MRID,
// 			COALESCE(MEASUREMENT_SIDE, '') AS MEASUREMENT_SIDE,
// 			COALESCE(AUTO_INTRODUCE_RULE, '') AS AUTO_INTRODUCE_RULE,
// 			COALESCE(OP_EXCHANGE, '') AS OP_EXCHANGE,
// 			COALESCE(AUTO_ALIAS, '') AS AUTO_ALIAS,
// 			COALESCE(CASE IMPORT_UPDATE_PROTECTION WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS IMPORT_UPDATE_PROTECTION,
// 			COALESCE(CASE IS_TEMP_SCADA WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS IS_TEMP_SCADA,
// 			COALESCE(STUDY_TOOLTIP_NAME, '') AS STUDY_TOOLTIP_NAME
// 		FROM COMPONENT_CLASS_DEFN
// 		WHERE COMPONENT_CLASS_INDEX = ?`, componentClassIndex)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &componentClassDefns, nil
// }

func readComponentClasses(db *sqlx.DB) ([]*ComponentClassDefn, error) {
	var componentClassDefns []*ComponentClassDefn

	err := db.Select(&componentClassDefns, `
		SELECT 
			COALESCE(COMPONENT_CLASS_INDEX, 0) AS COMPONENT_CLASS_INDEX,
			COALESCE(COMPONENT_CLASS_NAME, '') AS COMPONENT_CLASS_NAME,
			COALESCE(COMPONENT_STATUS, '') AS COMPONENT_STATUS,
			COALESCE(COMPONENT_ABBREVIATION, '') AS COMPONENT_ABBREVIATION,
			COALESCE(COMPONENT_APPEARANCE, '') AS COMPONENT_APPEARANCE,
			COALESCE(COMPONENT_LIFE_CYCLE, '') AS COMPONENT_LIFE_CYCLE,
			COALESCE(COMPONENT_EMS_CLASS_INDEX, 0) AS COMPONENT_EMS_CLASS_INDEX,
			COALESCE(COMPONENT_TRACE_COMPONENT, '') AS COMPONENT_TRACE_COMPONENT,
			COALESCE(COMPONENT_TRACE_LINE, '') AS COMPONENT_TRACE_LINE,
			COALESCE(CASE COMPONENT_IS_JUNCTION WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_JUNCTION,
			COALESCE(CASE COMPONENT_HAS_CUSTOMERS WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_HAS_CUSTOMERS,
			COALESCE(COMPONENT_NAME_RULE, '') AS COMPONENT_NAME_RULE,
			COALESCE(CASE COMPONENT_IS_SUP_INFEED WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_SUP_INFEED,
			COALESCE(CASE COMPONENT_IS_FEEDER_EQUIV WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_FEEDER_EQUIV,
			COALESCE(CASE COMPONENT_IS_ASSET WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_ASSET,
			COALESCE(CASE COMPONENT_IS_LOCATION WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_LOCATION,
			COALESCE(CASE COMPONENT_IS_TRANSFER_ATTR WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_TRANSFER_ATTR,
			COALESCE(CASE COMPONENT_IS_TRANSFER_ALIAS WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_TRANSFER_ALIAS,
			COALESCE(CASE COMPONENT_IS_TRANSFER_NAME WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_TRANSFER_NAME,
			COALESCE(CASE COMPONENT_IS_TRANSFER_PARENT WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_TRANSFER_PARENT,
			COALESCE(CASE NON_PATCHABLE WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS NON_PATCHABLE,
			COALESCE(COMPONENT_RT_CLASS, '') AS COMPONENT_RT_CLASS,
			COALESCE(CASE COMPONENT_DEL_ZONE_SHARABLE WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_DEL_ZONE_SHARABLE,
			COALESCE(CASE COMPONENT_IS_LV_RELEVANT WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_LV_RELEVANT,
			COALESCE(CASE COMPONENT_IS_TELEMETERED WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_TELEMETERED,
			COALESCE(COMPONENT_SLD_CLASS, '') AS COMPONENT_SLD_CLASS,
			COALESCE(APPLY_TO_CONN_COMP_CLASSES, '') AS APPLY_TO_CONN_COMP_CLASSES,
			COALESCE(TRACED_NAME_RULE, '') AS TRACED_NAME_RULE,
			COALESCE(TRACED_NAMING_PRIORITY, 0) AS TRACED_NAMING_PRIORITY,
			COALESCE(TRACED_NAMING_COMP_NAME_RULE, '') AS TRACED_NAMING_COMP_NAME_RULE,
			COALESCE(COMPONENT_CATEGORY, '') AS COMPONENT_CATEGORY,
			COALESCE(MENU_NAME, '') AS MENU_NAME,
			COALESCE(CASE SHOW_IN_EE_LOCATION_MODE WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS SHOW_IN_EE_LOCATION_MODE,
			COALESCE(ISOLATION_CLASS, '') AS ISOLATION_CLASS,
			COALESCE(CASE COMPONENT_IS_MAINTAINABLE WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_MAINTAINABLE,
			COALESCE(CASE CHECK_PARALLEL WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS CHECK_PARALLEL,
			COALESCE(MIXED_PHASE_STATES_SYMBOL, '') AS MIXED_PHASE_STATES_SYMBOL,
			COALESCE(CASE COMPONENT_IS_LV_SWITCH WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS COMPONENT_IS_LV_SWITCH,
			COALESCE(TRACE_CLASS_LOOKUP, '') AS TRACE_CLASS_LOOKUP,
			COALESCE(TOOLTIP_NAME, '') AS TOOLTIP_NAME,
			COALESCE(CASE IS_MULTIPLE_POSITION_SWITCH WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS IS_MULTIPLE_POSITION_SWITCH,
			COALESCE(CASE EXCLUDE_FROM_DEL_CAND WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS EXCLUDE_FROM_DEL_CAND,
			COALESCE(GENERATE_CIM_MRID, '') AS GENERATE_CIM_MRID,
			COALESCE(MEASUREMENT_SIDE, '') AS MEASUREMENT_SIDE,
			COALESCE(AUTO_INTRODUCE_RULE, '') AS AUTO_INTRODUCE_RULE,
			COALESCE(OP_EXCHANGE, '') AS OP_EXCHANGE,
			COALESCE(AUTO_ALIAS, '') AS AUTO_ALIAS,
			COALESCE(CASE IMPORT_UPDATE_PROTECTION WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS IMPORT_UPDATE_PROTECTION,
			COALESCE(CASE IS_TEMP_SCADA WHEN 'Y' THEN true WHEN 'N' THEN false ELSE NULL END, false) AS IS_TEMP_SCADA,
			COALESCE(STUDY_TOOLTIP_NAME, '') AS STUDY_TOOLTIP_NAME
		FROM COMPONENT_CLASS_DEFN`)
	if err != nil {
		return nil, err
	}
	return componentClassDefns, nil
}

func NewComponentClassDefns() *ComponentClassDefns {
	return &ComponentClassDefns{
		classDefByName:  make(map[string]*ComponentClassDefn),
		classDefByIndex: make(map[ComponentClassIndex]*ComponentClassDefn),
	}
}

var componentClassDefinitions *ComponentClassDefns // global used by String on ComponentClassIndex

func SetComponentClassDefinitions(cd *ComponentClassDefns) {
	componentClassDefinitions = cd
}

func GetComponentClasses(db *sqlx.DB) (*ComponentClassDefns, error) {
	// Change to a slice to hold multiple component class definitions
	componentClassDefnList, err := readComponentClasses(db)
	if err != nil {
		return nil, err
	}

	componentClassDefns := NewComponentClassDefns()

	for _, classDefn := range componentClassDefnList {
		componentClassDefns.classDefByName[classDefn.ComponentClassName] = classDefn
		componentClassDefns.classDefByIndex[classDefn.ComponentClassIndex] = classDefn
	}

	// Store the results in n
	return componentClassDefns, nil // Assuming n has a field to hold this data

}

func (c *ComponentClassDefns) GetComponentClassDefn(componentClassName string) (*ComponentClassDefn, error) {
	classDefn, ok := c.classDefByName[componentClassName]
	if !ok {
		return nil, fmt.Errorf("component class %s not found", componentClassName)
	}
	return classDefn, nil
}

func (c *ComponentClassDefns) GetComponentClassDefnByIndex(componentClassIndex ComponentClassIndex) (*ComponentClassDefn, error) {
	classDefn, ok := c.classDefByIndex[componentClassIndex]
	if !ok {
		return nil, fmt.Errorf("component class index %d not found", componentClassIndex)
	}
	return classDefn, nil
}

func (n *ComponentDb) GetComponentClassDetails(alias string) (*ComponentClassDetails, error) {

	comp, err := n.GetComponent(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting component: %w", err)
	}
	compClassDefn, err := n.GetComponentClassDefnByIndex(comp.ComponentClass)
	if err != nil {
		return nil, fmt.Errorf("error getting component class definition for comp %s : %w", alias, err)
	}

	return &ComponentClassDetails{ClassName: compClassDefn.ComponentClassName, NameRule: compClassDefn.ComponentNameRule, SubstationClass: comp.ComponentSubstationClass}, nil
}
