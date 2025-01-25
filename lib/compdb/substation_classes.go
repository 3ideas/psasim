package compdb

import "fmt"

type SubstationType int

const (
	NotApplicable SubstationType = iota
	PrimarySubstation
	SecondarySubstation
	PrimarySubstationComponent
	SecondarySubstationComponent
	PrimaryCircuitID
	SecondaryCircuitID
	LocationHolder
	PrimarySwitchgearSite
	SecondarySwitchgearSite
	PrimaryMinorSite
	SecondaryMinorSite
	PrimaryPanel
	SecondaryPanel
	PrimaryBusbar
	SecondaryBusbar
	PrimaryCircuitLocal
	SecondaryCircuitLocal
	PrimaryMainlineCircuit
	SecondaryMainlineCircuit
	PrimaryCircuit
	SecondaryCircuit
	PrimaryBay
	SecondaryBay
	LoadArea1TopArea
	LoadArea2SubArea
	LoadArea3ConformLoadGroup
	LoadArea3NonConformLoadGroup
)

var substationClassNames = []string{
	"Not Applicable", "Primary Substation", "Secondary Substation", "Primary Substation Component", "Secondary Substation Component", "Primary Circuit ID", "Secondary Circuit ID", "Location Holder", "Primary Switchgear Site", "Secondary Switchgear Site", "Primary Minor Site", "Secondary Minor Site", "Primary Panel", "Secondary Panel", "Primary Busbar", "Secondary Busbar", "Primary Circuit Local", "Secondary Circuit Local", "Primary Mainline Circuit", "Secondary Mainline Circuit", "Primary Circuit", "Secondary Circuit", "Primary Bay", "Secondary Bay", "Load Area 1 Top Area", "Load Area 2 Sub Area", "Load Area 3 Conform Load Group", "Load Area 3 Non Conform Load Group",
}

func (s SubstationType) String() string {
	if s < NotApplicable || s > LoadArea3NonConformLoadGroup {
		return "Unknown"
	}
	return substationClassNames[s]
}

func (s SubstationType) IsSubstation() bool {
	return s == PrimarySubstation || s == SecondarySubstation
}

func (s SubstationType) IsCircuit() bool {
	return s == PrimaryCircuitID || s == SecondaryCircuitID || s == PrimaryMainlineCircuit || s == SecondaryMainlineCircuit || s == PrimaryCircuit || s == SecondaryCircuit
}

func (s SubstationType) IsPrimaryCircuit() bool {
	return s == PrimaryCircuitID
}

func (s SubstationType) IsPlant() bool {
	return s == PrimarySubstationComponent || s == SecondarySubstationComponent
}

func (s SubstationType) IsComponent() bool { // TODO: check this
	//return !(s.IsSubstation() || s.IsCircuit())
	// return s == NotApplicable  // breaks some of the sames so indicating the text in the man page is that the substatin class is not applicatble not that it contains "Not Applicable"
	return true // TODO: check this
}

func GetSubstationClassFromName(substationClassName string) (SubstationType, error) {
	switch substationClassName {
	case "Primary Substation":
		return PrimarySubstation, nil
	case "Secondary Substation":
		return SecondarySubstation, nil
	case "Primary Substation Component":
		return PrimarySubstationComponent, nil
	case "Primary Circuit ID":
		return PrimaryCircuitID, nil
	case "Secondary Circuit ID":
		return SecondaryCircuitID, nil
	case "Location Holder":
		return LocationHolder, nil
	case "Primary Switchgear Site":
		return PrimarySwitchgearSite, nil
	case "Secondary Switchgear Site":
		return SecondarySwitchgearSite, nil
	case "Primary Minor Site":
		return PrimaryMinorSite, nil
	case "Secondary Minor Site":
		return SecondaryMinorSite, nil
	case "Primary Panel":
		return PrimaryPanel, nil
	case "Secondary Panel":
		return SecondaryPanel, nil
	case "Primary Busbar":
		return PrimaryBusbar, nil
	case "Secondary Busbar":
		return SecondaryBusbar, nil
	case "Primary Circuit Local":
		return PrimaryCircuitLocal, nil
	case "Secondary Circuit Local":
		return SecondaryCircuitLocal, nil
	case "Primary Mainline Circuit":
		return PrimaryMainlineCircuit, nil
	case "Secondary Mainline Circuit":
		return SecondaryMainlineCircuit, nil
	case "Primary Circuit":
		return PrimaryCircuit, nil
	case "Secondary Circuit":
		return SecondaryCircuit, nil
	case "Primary Bay":
		return PrimaryBay, nil
	case "Secondary Bay":
		return SecondaryBay, nil
	case "Load Area 1 Top Area":
		return LoadArea1TopArea, nil
	case "Load Area 2 Sub Area":
		return LoadArea2SubArea, nil
	case "Load Area 3 Conform Load Group":
		return LoadArea3ConformLoadGroup, nil
	case "Load Area 3 Non Conform Load Group":
		return LoadArea3NonConformLoadGroup, nil
	default:
		return NotApplicable, fmt.Errorf("unknown substation class: %s", substationClassName)
	}
}
