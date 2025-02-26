package psalerts

import (
	"fmt"
	"log/slog"

	"github.com/3ideas/psasim/lib/namerif"
)

type AliasResolver interface {
	LookupAlias(alias string) (string, bool)
}

func (p *PSAlerts) ResolveAliases(ns namerif.NameService, eterraAliasResolver AliasResolver, resolvedAlarmsFile string, unresolvedAlarmsFile string) error {

	count := 0
	resolved := 0
	notFound := 0

	resolvedAlerts := &PSAlerts{}
	unresolvedAlerts := &PSAlerts{}
	// var ok bool
	for _, alert := range p.Alerts {

		if alert.AlarmComponentAlias == "OAKLANDS CRESCENT S-RTU" {
			fmt.Printf("Alert %d: %+v\n", alert.LineNumber, alert)
		}
		count++

		poAliasInfo, err := ns.GetComponentInfo(alert.Alias)
		if err != nil && eterraAliasResolver != nil {
			poAlias, ok := eterraAliasResolver.LookupAlias(alert.Alias)
			if !ok {
				unresolvedAlerts.Add(alert)
				slog.Warn("Alias not found", "alias", alert.Alias, "Original alias", alert.AlarmComponentAlias, "type", alert.AlarmType, "substation alias", alert.AlarmSubstationAlias, "substation name", alert.AlarmSubstationName,
					"alarm name", alert.AlarmName, "alarm text", alert.AlarmText, "alarm text2", alert.AlarmText2,
					"descriptor", alert.Descriptor, "component pathname", alert.ComponentPathname, "primary busbar", alert.PrimaryBusbar, "primary feeder", alert.PrimaryFeeder, "lineNo", alert.LineNumber)
				notFound++
				continue
			} else {
				if poAlias == "" {
					slog.Info("Eterra entry found but no mapping to PO alias!", "alias", alert.Alias, "Original alias", alert.AlarmComponentAlias, "type", alert.AlarmType, "substation alias", alert.AlarmSubstationAlias, "substation name", alert.AlarmSubstationName,
						"alarm name", alert.AlarmName, "alarm text", alert.AlarmText, "alarm text2", alert.AlarmText2,
						"descriptor", alert.Descriptor, "component pathname", alert.ComponentPathname, "primary busbar", alert.PrimaryBusbar, "primary feeder", alert.PrimaryFeeder, "lineNo", alert.LineNumber)
					notFound++
					unresolvedAlerts.Add(alert)
					continue
				} else {
					poAliasInfo, err = ns.GetComponentInfo(poAlias)
					if err != nil {
						slog.Error("Alias not found in PO after eterra/po mapping was found!", "alias", alert.Alias, "Original alias", alert.AlarmComponentAlias, "type", alert.AlarmType, "substation alias", alert.AlarmSubstationAlias, "substation name", alert.AlarmSubstationName,
							"alarm name", alert.AlarmName, "alarm text", alert.AlarmText, "alarm text2", alert.AlarmText2,
							"descriptor", alert.Descriptor, "component pathname", alert.ComponentPathname, "primary busbar", alert.PrimaryBusbar, "primary feeder", alert.PrimaryFeeder, "lineNo", alert.LineNumber)
						notFound++
						unresolvedAlerts.Add(alert)
						continue
					}
				}
			}
		}
		poAlias := poAliasInfo.Alias

		alert.Alias = poAlias

		resolvedAlerts.Add(alert)

		slog.Info("Alias found", "alias", alert.Alias, "type", alert.AlarmType, "alert", alert)
		resolved++
	}

	// Write the resolved alarms to the file as a CSV
	resolvedAlerts.WriteCSV(resolvedAlarmsFile)
	if unresolvedAlarmsFile != "" {
		unresolvedAlerts.WriteCSV(unresolvedAlarmsFile)
	}
	fmt.Printf("Resolved %d of %d alarms\n", resolved, count)
	fmt.Printf("Not found %d of %d alarms\n", notFound, count)

	return nil
}
