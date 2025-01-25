package psalerts

import (
	"fmt"
	"log/slog"

	"github.com/3ideas/psasim/lib/namerif"
)

type AliasResolver interface {
	LookupAlias(alias string) (string, bool)
}

func (p *PSAlerts) ResolveAliases(ns namerif.NameService, eterraAliasResolver AliasResolver) error {

	count := 0
	resolved := 0
	notFound := 0
	for _, alarm := range p.Alerts {
		count++

		_, err := ns.GetComponentInfo(alarm.Alias)
		if err != nil && eterraAliasResolver != nil {
			poAlias, ok := eterraAliasResolver.LookupAlias(alarm.Alias)
			if !ok {
				slog.Warn("Alias not found", "alias", alarm.Alias, "Original alias", alarm.AlarmComponentAlias, "type", alarm.AlarmType, "substation alias", alarm.AlarmSubstationAlias, "substation name", alarm.AlarmSubstationName,
					"alarm name", alarm.AlarmName, "alarm text", alarm.AlarmText, "alarm text2", alarm.AlarmText2,
					"descriptor", alarm.Descriptor, "component pathname", alarm.ComponentPathname, "primary busbar", alarm.PrimaryBusbar, "primary feeder", alarm.PrimaryFeeder)
				notFound++
				continue
			} else {
				_, err := ns.GetComponentInfo(poAlias)
				if err != nil {
					slog.Warn("Alias not found", "alias", alarm.Alias, "Original alias", alarm.AlarmComponentAlias, "type", alarm.AlarmType, "substation alias", alarm.AlarmSubstationAlias, "substation name", alarm.AlarmSubstationName,
						"alarm name", alarm.AlarmName, "alarm text", alarm.AlarmText, "alarm text2", alarm.AlarmText2,
						"descriptor", alarm.Descriptor, "component pathname", alarm.ComponentPathname, "primary busbar", alarm.PrimaryBusbar, "primary feeder", alarm.PrimaryFeeder)
					notFound++
					continue
				}
			}
		}
		slog.Info("Alias found", "alias", alarm.Alias, "type", alarm.AlarmType, "alert", alarm)
		resolved++
	}

	fmt.Printf("Resolved %d of %d alarms\n", resolved, count)
	fmt.Printf("Not found %d of %d alarms\n", notFound, count)

	return nil
}
