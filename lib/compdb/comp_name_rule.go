package compdb

import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/jmoiron/sqlx"
)

type TextLocationType int

const (
	NameRuleLocation TextLocationType = 1
	NameRuleCircuit  TextLocationType = 2
	NameRulePlant    TextLocationType = 3
	NameRuleOrigin   TextLocationType = 4
	NameRuleParent   TextLocationType = 5
)

func (t TextLocationType) String() string {
	switch t {
	case NameRuleLocation:
		return "Location"
	case NameRuleCircuit:
		return "Circuit"
	case NameRulePlant:
		return "Plant"
	case NameRuleOrigin:
		return "Origin"
	case NameRuleParent:
		return "Parent"
	default:
		return "Unknown"
	}
}

type TextTypeType int

const (
	NameRuleTextTypeAttributeElseName                  TextTypeType = 1
	NameRuleTextTypeAttributeValue                     TextTypeType = 2
	NameRuleTextTypeAbbriviation                       TextTypeType = 3
	NameRuleTextTypeData                               TextTypeType = 4
	NameRuleTextTypeForceNextInstruction               TextTypeType = 5
	NameRuleTextTypeVoltage                            TextTypeType = 6
	NameRuleTextTypeAlias                              TextTypeType = 7
	NameRuleTextTypeUserReference                      TextTypeType = 8
	NameRuleTextTypeAttributeOrigin                    TextTypeType = 9  // this is for the default rule - not a PO rule
	NameRuleTextTypeAttributeOriginIfDifferentElseName TextTypeType = 10 // this is for the default rule - not a PO rule
)

func (t TextTypeType) String() string {
	switch t {
	case NameRuleTextTypeAttributeElseName:
		return "Attribute Value ELSE Network Device Name"
	case NameRuleTextTypeAttributeValue:
		return "Attribute Value"
	case NameRuleTextTypeAbbriviation:
		return "Abbreviation"
	case NameRuleTextTypeData:
		return "Text specified in \"Data\""
	case NameRuleTextTypeForceNextInstruction:
		return "Force next instruction to this component..."
	case NameRuleTextTypeVoltage:
		return "Voltage"
	case NameRuleTextTypeAlias:
		return "Alias"
	case NameRuleTextTypeUserReference:
		return "User Reference"
	case NameRuleTextTypeAttributeOrigin:
		return "Attribute Origin"
	case NameRuleTextTypeAttributeOriginIfDifferentElseName:
		return "Origin Attribute Value IF Different from circuit ELSE Origin Name"
	default:
		return "Unknown"
	}
}

type ComponentNameRule struct {
	NameRule            string           `db:"NAME_RULE"`
	TextIndex           int              `db:"TEXT_INDEX"`
	TextLocation        TextLocationType `db:"TEXT_LOCATION"`
	TextType            TextTypeType     `db:"TEXT_TYPE"`
	Data                string           `db:"DATA"`
	PreText             string           `db:"PRE_TEXT"`
	PostText            string           `db:"POST_TEXT"`
	Comments            string           `db:"COMMENTS"`
	Data2               string           `db:"DATA2"`
	UseSeparator        bool             `db:"USE_SEPARATOR"`
	UseParentIfNotFound bool             `db:"USE_PARENT_IF_NOT_FOUND"` // These are not PO rules
	UseIfNotFound       bool             `db:"USE_IF_NOT_FOUND"`        // These are not PO rules
	UseOriginIfNotFound bool             `db:"USE_ORIGIN_IF_NOT_FOUND"` // These are not PO rules
}

type ComponentNameRules struct {
	nameRules map[string][]*ComponentNameRule
}

func NewComponentNameRules() *ComponentNameRules {
	return &ComponentNameRules{
		nameRules: make(map[string][]*ComponentNameRule),
	}
}

func (c *ComponentNameRule) String() string {
	return fmt.Sprintf("NameRule: %s, TextIndex: %d, TextLocation: %s, TextType: %s, Data: %s, PreText: %s, PostText: %s, Comments: %s, Data2: %s, UseSeparator: %t", c.NameRule, c.TextIndex, c.TextLocation, c.TextType, c.Data, c.PreText, c.PostText, c.Comments, c.Data2, c.UseSeparator)
}

// GetComponentNameRulesList
// Not use seperator is defaulted to true if null.
func GetComponentNameRulesList(db *sqlx.DB) ([]*ComponentNameRule, error) {
	var componentNameRules []*ComponentNameRule
	err := db.Select(&componentNameRules, `
		SELECT 
			COALESCE(NAME_RULE, '') AS NAME_RULE,
			COALESCE(TEXT_INDEX, 0) AS TEXT_INDEX,
			COALESCE(TEXT_LOCATION, '') AS TEXT_LOCATION,
			COALESCE(TEXT_TYPE, '') AS TEXT_TYPE,
			COALESCE(DATA, '') AS DATA,
			COALESCE(PRE_TEXT, '') AS PRE_TEXT,
			COALESCE(POST_TEXT, '') AS POST_TEXT,
			COALESCE(COMMENTS, '') AS COMMENTS,
			COALESCE(DATA2, '') AS DATA2,
			COALESCE(CASE USE_SEPARATOR WHEN '1' THEN true WHEN '0' THEN false ELSE NULL END, true) AS USE_SEPARATOR
		FROM COMPONENT_NAME_RULE 
		 ORDER BY NAME_RULE,TEXT_INDEX`)
	if err != nil {
		return nil, err
	}
	return componentNameRules, nil
}

func GetComponentNameRules(db *sqlx.DB) (*ComponentNameRules, error) {

	componentNameRules := NewComponentNameRules()

	nameRules, err := GetComponentNameRulesList(db)
	if err != nil {
		return nil, err
	}
	for _, nameRule := range nameRules {
		componentNameRules.nameRules[nameRule.NameRule] = append(componentNameRules.nameRules[nameRule.NameRule], nameRule)
	}

	return componentNameRules, nil

}

func (c *ComponentNameRules) GetComponentNameRule(nameRule string) ([]*ComponentNameRule, bool) {
	rules, ok := c.nameRules[nameRule]
	return rules, ok
}

// GetAllNameRules return a list of distinct name rules
func (c *ComponentNameRules) GetAllNameRules() []string {
	rules := make([]string, 0, len(c.nameRules))
	ruleSet := make(map[string]struct{})
	for nameRule := range c.nameRules {
		if _, found := ruleSet[nameRule]; !found {
			ruleSet[nameRule] = struct{}{}
			rules = append(rules, nameRule)
		}
	}
	// Sort the rules
	sort.Strings(rules)
	return rules
}

func (n *ComponentDb) GetNameRulesForComponents(componentAliases []string) ([]string, error) {

	nameRules := make(map[string]struct{})
	for _, alias := range componentAliases {
		ruleName, rules, err := n.GetNameRulesForComponent(alias)
		if err != nil {
			if ruleName != "" {
				slog.Info("GetNameRulesForComponents: unable to get name rules", "Component", alias, "error", err)
			}
			continue
		}

		nameRules[rules[0].NameRule] = struct{}{}
	}

	distinctRules := make([]string, 0, len(nameRules))
	for rule := range nameRules {
		distinctRules = append(distinctRules, rule)
	}
	sort.Strings(distinctRules)
	return distinctRules, nil
}
