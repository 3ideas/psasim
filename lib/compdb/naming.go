package compdb

import (
	"fmt"
	"regexp"
	"strings"
)

type NameLocationType int

const (
	NLPathName          NameLocationType = 1
	NLAlias             NameLocationType = 2
	NLAttrubute         NameLocationType = 3
	NLClassAbbreviation NameLocationType = 4
	NLAttrubuteOrigin   NameLocationType = 5
	NLNameOrigin        NameLocationType = 6
)

type NamePartDetails struct {
	Value         string
	Comp          *Component
	LocationType  NameLocationType
	AttributeName string
	RawValue      string
	Separator     string
}

type NamePart struct {
	Name    string
	Details []*NamePartDetails
}

func (n NamePart) String() string {
	multipleDetails := ""
	if len(n.Details) > 1 {
		multipleDetails = fmt.Sprintf(" (%d)", len(n.Details))
	}
	return fmt.Sprintf("Name: %s%s", n.Name, multipleDetails)
}

// This is used for interface use (no pointers to comps etc)
// it mirrors what the rpc interface uses
type NameDetails struct {
	Name     string
	RuleName string
	Alias    string
	Pathname string
	Location *NamePartResponse
	Circuit  *NamePartResponse
	Plant    *NamePartResponse
	Origin   *NamePartResponse
}

func (n NameDetails) String() string {
	return fmt.Sprintf("Name: %15s  Rule: %-15s  Location: %-20s  Circuit: %-20s Plant: %-20s  Origin: %-20s", n.Name, n.RuleName, n.Location.Value, n.Circuit.Value, n.Plant.Value, n.Origin.Value)
}

// This is used for interface use, it mirrors what the rpc interface uses
type NamePartResponse struct {
	Value   string
	Alias   string
	Used    bool
	Details []*NamePartDetailsResponse
}

func (n NamePartResponse) String() string {
	multipleDetails := ""
	for _, detail := range n.Details {
		multipleDetails += fmt.Sprintf("\n             Val: %-15s: Src: %-20s, Raw: %-15s  Sep: %3s", detail.Value, detail.Source, detail.RawValue, detail.Separator)
	}
	return fmt.Sprintf("Name: %-15s Comp: %-35s Used: %t%s", n.Value, n.Alias, n.Used, multipleDetails)
}

type NamePartDetailsResponse struct {
	Value     string
	Source    string
	RawValue  string
	Separator string
}

type NameDetailsFull struct {
	NameDetails
	Rule    []*ComponentNameRule
	Parents []*Component
}

func (n NameDetailsFull) String() string {
	return n.NameDetails.String()
}

func (n NameDetailsFull) OriginComponent() *Component {
	if len(n.Parents) > 0 {
		return n.Parents[0]
	}
	return nil
}

func (n *ComponentDb) GetNameFull(alias string) (*NameDetailsFull, error) {

	ruleName, nameRules, err := n.GetNameRulesForComponent(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting name rules for component %s, %s, %w", alias, ruleName, err)
	}

	parents, err := n.GetParents(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting parents for component %s, %w", alias, err)
	}

	name := n.getNameFromRules(parents, nameRules)
	return name, nil
}

func (n *ComponentDb) GetName(alias string) (*NameDetails, error) {

	name, err := n.GetNameFull(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting name for component %s, %w", alias, err)
	}

	return &name.NameDetails, nil
}

// ResolveNames resolves the names for all components
// It sets the Name field for each component
// It also builds the componentsByName map
func (n *ComponentDb) ResolveNames() {

	for _, component := range n.Components.componentsByAlias {

		name, err := n.GetName(component.ComponentAlias)
		if err != nil {
			continue
		}
		if name.Name == "" {
			continue
		}
		component.Name = name.Name
	}
	n.Components.BuildComponentByName()
}

func (n *ComponentDb) getNameFromRules(parents []*Component, nameRules []*ComponentNameRule) *NameDetailsFull {

	if len(nameRules) == 0 {
		nameRules = DefaultNameRules()
	}

	if nameRules[0].TextLocation == NameRuleParent {
		nameRules = getNameRulesForParent(nameRules)
	}

	var nameFull NameDetailsFull

	location := n.getPartName(nameRules, parents, NameRuleLocation)
	circuit := n.getPartName(nameRules, parents, NameRuleCircuit)
	plant := n.getPartName(nameRules, parents, NameRulePlant)
	origin := n.getPartName(nameRules, parents, NameRuleOrigin)

	if location.Name == "" {
		location.Name = parents[0].ComponentAlias
		location.Details = append(location.Details, &NamePartDetails{
			Comp:          parents[0],
			LocationType:  NLAlias,
			AttributeName: "",
			Value:         parents[0].ComponentAlias,
		})
	}
	nameFull.Alias = parents[0].ComponentAlias
	nameFull.Pathname = parents[0].ComponentPathname
	nameFull.Rule = nameRules
	nameFull.RuleName = nameRules[len(nameRules)-1].NameRule // Use the last rule for the name rule as if we added additional rules to the start they don't change the name rule name
	nameFull.Location = n.convertNamePart(location)
	nameFull.Circuit = n.convertNamePart(circuit)
	nameFull.Plant = n.convertNamePart(plant)
	nameFull.Origin = n.convertNamePart(origin)
	nameFull.Parents = parents

	// Join circuit and plant and origin, if plant is not empty do not add origin
	// use space as separator
	var circuitAndCompName string

	circuitNameUsed := false
	plantNameUsed := false
	originNameUsed := false

	if circuit.Name != "" {
		circuitAndCompName = circuit.Name
		circuitNameUsed = true
	}

	if plant.Name != "" {
		if circuitAndCompName != "" {
			circuitAndCompName += " "
		}
		plantNameUsed = true
		circuitAndCompName += plant.Name
	}

	if plant.Name == "" && origin.Name != "" {
		if circuitAndCompName != "" {
			circuitAndCompName += " "
		}
		circuitAndCompName += origin.Name
		originNameUsed = true
	}

	if circuitAndCompName == "" {
		nameFull.Name = location.Name
		nameFull.Location.Used = true
		nameFull.Circuit.Used = false
		nameFull.Plant.Used = false
		nameFull.Origin.Used = false
	} else {
		nameFull.Name = location.Name + ", " + circuitAndCompName
		nameFull.Location.Used = true
		nameFull.Circuit.Used = circuitNameUsed
		nameFull.Plant.Used = plantNameUsed
		nameFull.Origin.Used = originNameUsed
	}

	return &nameFull
}

// convertNamePart converts a NamePart to a NamePartResponse
// Note this does not populat the Used flag
func (n *ComponentDb) convertNamePart(namePart *NamePart) *NamePartResponse {
	if len(namePart.Details) == 0 {
		return &NamePartResponse{}
	}
	namePartResponse := &NamePartResponse{Value: namePart.Name, Alias: namePart.Details[0].Comp.ComponentAlias}
	for _, detail := range namePart.Details {
		source := n.getSourceDescription(detail)
		namePartResponse.Details = append(namePartResponse.Details, &NamePartDetailsResponse{Value: detail.Value, Source: source, RawValue: detail.RawValue, Separator: detail.Separator})
	}
	return namePartResponse
}

func (n *ComponentDb) getSourceDescription(detail *NamePartDetails) string {
	var source string
	if detail.LocationType == NLAlias {
		source = "Alias: " + detail.Comp.ComponentAlias
	} else if detail.LocationType == NLPathName {
		source = "Path: " + detail.Comp.ComponentPathname
	} else if detail.LocationType == NLAttrubute {
		source = "Attr: " + detail.AttributeName
	} else if detail.LocationType == NLClassAbbreviation {
		classDefn, err := n.GetComponentClassDefnByIndex(detail.Comp.ComponentClass)
		if err != nil {
			source = "Class: " + fmt.Sprintf("%d - Unknown", detail.Comp.ComponentClass)
		} else {
			source = "Class: " + classDefn.ComponentClassName
		}
	} else if detail.LocationType == NLAttrubuteOrigin {
		source = "AttrOrigin: " + detail.AttributeName
	} else {
		source = "Unknown"
	}
	return source
}

func (n *ComponentDb) getPartName(ruleSet []*ComponentNameRule, parents []*Component, textLocation TextLocationType) *NamePart {

	rules := getPartRules(ruleSet, textLocation)

	var forceComponent *Component
	partName := ""
	var namePartDetails []*NamePartDetails
	for _, rule := range rules {
		var namePartDetail *NamePartDetails

		namePartDetail, forceComponent = n.getNameForRule(parents, rule, forceComponent)
		if namePartDetail == nil {
			continue
		}
		if rule.UseIfNotFound && partName != "" {
			continue
		}
		if partName != "" { // TODO: check if there is a space on the end currently and not add another one
			partName += " "
		}

		partName += namePartDetail.Value
		namePartDetails = append(namePartDetails, namePartDetail)

	}

	return &NamePart{partName, namePartDetails}
}

// getNameRulesForParent adds rules to the name rules list this is used
// when the first rule in the list is a NameRuleParent(5)
// it adds rules for getting the location Pathname,  Abbreviation and a fallback rule to get the alias of the origin if the location is not found
// it also adds a rule to get the circuit name
func getNameRulesForParent(nameRules []*ComponentNameRule) []*ComponentNameRule {

	locationAttrOrNameRule := &ComponentNameRule{
		TextLocation: NameRuleLocation,
		TextType:     NameRuleTextTypeAttributeElseName,
		Data:         "Not Applicable",
		PostText:     "",
	}

	locationAbbreviationRule := &ComponentNameRule{
		TextLocation: NameRuleLocation,
		TextType:     NameRuleTextTypeAbbriviation,
		Data:         "Not Applicable",
	}

	// locationAliasOriginRule := &ComponentNameRule{
	// 	TextLocation:  NameRuleLocation,
	// 	TextType:      NameRuleTextTypeAliasOrigin,
	// 	Data:          "Not Applicable",
	// 	UseIfNotFound: true,
	// }

	circuitNameRule := &ComponentNameRule{
		TextLocation:        NameRuleCircuit,
		TextType:            NameRuleTextTypeAttributeElseName,
		Data:                "Circuit Name",
		UseParentIfNotFound: true,
	}

	nameRules = append([]*ComponentNameRule{
		locationAttrOrNameRule,
		locationAbbreviationRule,
		// locationAliasOriginRule,
		circuitNameRule,
	}, nameRules...)

	return nameRules
}

func DefaultNameRules() []*ComponentNameRule {
	locationAttrOrNameRule := &ComponentNameRule{
		NameRule:     "<default>",
		TextLocation: NameRuleLocation,
		TextType:     NameRuleTextTypeAttributeElseName,
		Data:         "Not Applicable",
		PostText:     "",
	}

	locationAbbreviationRule := &ComponentNameRule{
		NameRule:     "<default>",
		TextLocation: NameRuleLocation,
		TextType:     NameRuleTextTypeAbbriviation,
		Data:         "Not Applicable",
	}

	// locationAliasOriginRule := &ComponentNameRule{
	// 	TextLocation:  NameRuleLocation,
	// 	TextType:      NameRuleTextTypeAliasOrigin,
	// 	Data:          "Not Applicable",
	// 	UseIfNotFound: true,
	// }

	circuitNameRule := &ComponentNameRule{
		NameRule:     "<default>",
		TextLocation: NameRuleCircuit,
		TextType:     NameRuleTextTypeAttributeElseName,
		Data:         "Circuit Name",
		// UseParentIfNotFound: true,
		UseOriginIfNotFound: true,
	}

	// circuitNameRule2 := &ComponentNameRule{
	// 	NameRule:            "<default>",
	// 	TextLocation:        NameRuleCircuit,
	// 	TextType:            NameRuleTextTypeAttributeOrigin,
	// 	Data:                "Circuit Name",
	// 	UseParentIfNotFound: true,
	// }

	// plantNameRule1 := &ComponentNameRule{
	// 	NameRule:            "<default>",
	// 	TextLocation:        NameRulePlant,
	// 	TextType:            NameRuleTextTypeAttributeValue,
	// 	Data:                "Plant", // Not
	// 	UseParentIfNotFound: true,
	// }

	originNameRule1 := &ComponentNameRule{
		NameRule:            "<default>",
		TextLocation:        NameRuleCircuit,
		TextType:            NameRuleTextTypeAttributeOriginIfDifferentElseName,
		Data:                "Circuit Name", // Not
		UseParentIfNotFound: true,
	}

	nameRules := []*ComponentNameRule{
		locationAttrOrNameRule,
		locationAbbreviationRule,
		// locationAliasOriginRule,
		circuitNameRule,
		// circuitNameRule2,
		// plantNameRule1,
		originNameRule1,
	}

	return nameRules
}

func getPartRules(nameRules []*ComponentNameRule, textLocation TextLocationType) []*ComponentNameRule {
	locationRules := []*ComponentNameRule{}
	for i, rule := range nameRules {

		if rule.TextLocation == textLocation {
			if i != 0 && nameRules[i-1].TextLocation == NameRuleParent { // if the previous rule is a parent rule then  add it to the list
				locationRules = append(locationRules, nameRules[i-1])
			}
			locationRules = append(locationRules, rule)
		}
	}
	return locationRules
}

func (n *ComponentDb) getNameForRule(parents []*Component, rule *ComponentNameRule, forceComponent *Component) (*NamePartDetails, *Component) {

	var comp *Component

	fallback := parents[0]

	if forceComponent == nil {
		comp = n.getComponentForRule(rule, parents)
		fallback = comp // TODO: Why is fallback not the origin by default?
	} else {
		// if we have a forced component then the fallback component is the origin...
		comp = forceComponent
	}

	if comp == nil {
		return nil, nil
	}

	namePartDetails, forceComponentFlag := n.getNameValue(rule, comp, fallback, parents[0])

	if forceComponentFlag {
		return nil, comp
	}

	if namePartDetails == nil {
		return nil, nil
	}

	if rule.PostText != "" {
		namePartDetails.Value += rule.PostText
		namePartDetails.Separator = rule.PostText
	}

	return namePartDetails, nil

}

func (n *ComponentDb) getNameValue(r *ComponentNameRule, comp *Component, fallback *Component, origin *Component) (*NamePartDetails, bool) {

	switch r.TextType {
	case NameRuleTextTypeAttributeElseName: //Attribute Value ELSE Network Device Name
		attr, err := n.GetComponentAttribute(comp.ComponentID, r.Data)
		if err == nil {
			if attr.AttributeValue == "" {
				return &NamePartDetails{
					Comp:          fallback,
					LocationType:  NLPathName,
					AttributeName: "",
					Value:         fallback.ComponentPathname,
					RawValue:      fallback.ComponentPathname,
				}, false
			}
			attrValue := attr.AttributeValue
			attrRawValue := attr.AttributeValue
			if r.Data == "State Alarm Text" {
				// Parse for string formatting
				attrValue = strings.ReplaceAll(attrValue, "%%", "%")                    // Replace "%%" with "%"
				attrValue = regexp.MustCompile(`%[^%]`).ReplaceAllString(attrValue, "") // Remove "%.."
			}
			return &NamePartDetails{
				Comp:          comp,
				LocationType:  NLAttrubute,
				AttributeName: r.Data,
				Value:         attrValue,
				RawValue:      attrRawValue,
			}, false
		}
		return &NamePartDetails{
			Comp:          fallback,
			LocationType:  NLPathName,
			AttributeName: "",
			Value:         fallback.ComponentPathname,
			RawValue:      fallback.ComponentPathname,
		}, false
	case NameRuleTextTypeAttributeValue: // Attribute Value
		attr, err := n.GetComponentAttribute(comp.ComponentID, r.Data)
		if err == nil {
			if attr.AttributeValue == "" {
				return nil, false
			}
			attrValue := attr.AttributeValue
			attrRawValue := attr.AttributeValue
			if r.Data == "State Alarm Text" || r.Data == "Supplementary Text" {
				// Parse for string formatting
				if strings.Contains(attrValue, "%% ") {
					attrValue = strings.ReplaceAll(attrValue, "%%", "%")
				} else {
					attrValue = regexp.MustCompile(`%[^%]`).ReplaceAllString(attrValue, "") // Remove "%.."
				} // Replace "%%" with "%"

			}
			return &NamePartDetails{
				Comp:          comp,
				LocationType:  NLAttrubute,
				AttributeName: r.Data,
				Value:         attrValue,
				RawValue:      attrRawValue,
			}, false
		}

		return nil, false

	case NameRuleTextTypeAbbriviation: // Abbreviation
		compClass, err := n.GetComponentClassDefnByIndex(comp.ComponentClass)
		if err != nil {
			return nil, false
		}
		return &NamePartDetails{
			Comp:          comp,
			LocationType:  NLClassAbbreviation,
			AttributeName: "",
			Value:         compClass.ComponentAbbreviation,
			RawValue:      compClass.ComponentAbbreviation,
		}, false

	case NameRuleTextTypeForceNextInstruction: // Force next instruction to this component...
		return nil, true

	case NameRuleTextTypeAlias:
		return &NamePartDetails{
			Comp:          comp,
			LocationType:  NLAlias,
			AttributeName: "",
			Value:         comp.ComponentAlias,
			RawValue:      comp.ComponentAlias,
		}, false

	case NameRuleTextTypeAttributeOrigin:
		attr, err := n.GetComponentAttribute(fallback.ComponentID, r.Data)
		if err == nil {
			if attr.AttributeValue == "" {
				return nil, false
			}
			return &NamePartDetails{
				Comp:          comp,
				LocationType:  NLAttrubuteOrigin,
				AttributeName: r.Data,
				Value:         attr.AttributeValue,
				RawValue:      attr.AttributeValue,
			}, false
		}

		return nil, false

	case NameRuleTextTypeAttributeOriginIfDifferentElseName:

		// Get the 1st part of the circuit name
		circuitNamePart1 := ""
		attr, err := n.GetComponentAttribute(comp.ComponentID, r.Data)
		if err == nil && attr.AttributeValue != "" {
			circuitNamePart1 = attr.AttributeValue
		} else {
			circuitNamePart1 = comp.ComponentPathname
		}

		//
		// Get the 2nd part of the circuit name
		originAttrValue := ""
		originAttr, err := n.GetComponentAttribute(origin.ComponentID, r.Data)
		if err == nil {
			originAttrValue = originAttr.AttributeValue
		}

		if circuitNamePart1 != originAttrValue && originAttrValue != "" { // if we got an origin attribute value and it is different from part 1 use it
			return &NamePartDetails{
				Comp:          origin,
				LocationType:  NLAttrubuteOrigin,
				AttributeName: r.Data,
				Value:         originAttrValue,
				RawValue:      originAttrValue,
			}, false
			// } else if !strings.HasPrefix(circuitNamePart1, origin.ComponentPathname) { // if the origin pathname is not a  prefix of part 1 use it {
			// 	return &NamePartDetails{
			// 		comp:          origin,
			// 		locationType:  NLNameOrigin,
			// 		attributeName: "",
			// 		Name:          origin.ComponentPathname,
			// 	}, false

		} else if comp.ComponentPathname != circuitNamePart1 { // if comp pathname different from 1st part use it
			return &NamePartDetails{
				Comp:          comp,
				LocationType:  NLPathName,
				AttributeName: "",
				Value:         comp.ComponentPathname,
				RawValue:      comp.ComponentPathname,
			}, false
		}
		// Last resort use the origin class abbreviation
		compClass, err := n.GetComponentClassDefnByIndex(origin.ComponentClass)
		if err != nil || compClass.ComponentAbbreviation == "" {
			return nil, false
		}
		return &NamePartDetails{
			Comp:          origin,
			LocationType:  NLClassAbbreviation,
			AttributeName: "",
			Value:         compClass.ComponentAbbreviation,
			RawValue:      compClass.ComponentAbbreviation,
		}, false

	default:
		fmt.Printf("TextType not implemented, %s  %v", r.TextType, r)
	}

	return nil, false
}

func (n *ComponentDb) getComponentForRule(r *ComponentNameRule, parents []*Component) *Component {
	switch r.TextLocation {
	case NameRuleLocation: // Location
		for _, comp := range parents { // skip 1st component when looking for location

			if comp.ComponentSubstationClass.IsSubstation() {
				return comp
			}
		}
		return nil

	case NameRuleCircuit: // Circuit
		for _, comp := range parents {
			// TODO: check if this is correct
			if comp.ComponentSubstationClass.IsCircuit() {
				return comp
			}

		}
		if r.UseOriginIfNotFound {
			return parents[0]
		}
		if len(parents) > 1 && r.UseParentIfNotFound { // && parents[1].ComponentSubstationClass == NotApplicable {
			return parents[1]
		}
		return nil

	case NameRulePlant: // Plant
		// fmt.Printf("Rule: %v \n", r)
		for _, comp := range parents {
			if comp.ComponentSubstationClass.IsPlant() {
				return comp
			}

		}
		return nil

	case NameRuleOrigin: // Origin
		if parents[0].ComponentSubstationClass.IsComponent() {
			return parents[0]
		}
		return nil

	case NameRuleParent: // Parent
		if len(parents) >= 2 {
			return parents[1]
		}
		return nil

	}
	return nil
}
