package compare

type EterraToPO struct {
	eTerraToPowerOn map[string]string
}

// EterraToPO generates a lookup object for  eterra aliases to PO ones
func (a *AlarmsComparison) EterraToPO() *EterraToPO {

	eTerraToPowerOnMap := make(map[string]string)

	for _, alarm := range a.Alarms() {
		eTerraToPowerOnMap[alarm.ETerraAlias] = alarm.POAlias
	}

	return &EterraToPO{eTerraToPowerOn: eTerraToPowerOnMap}
}

func (e *EterraToPO) LookupAlias(eterraAlias string) (string, bool) {
	poAlias, ok := e.eTerraToPowerOn[eterraAlias]
	return poAlias, ok
}
