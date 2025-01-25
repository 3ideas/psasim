package psalerts

import (
	"fmt"
	"log/slog"

	"github.com/3ideas/psasim/lib/namerif"
)

func (p *PSAlerts) ResolveAliases(ns namerif.NameService) error {

	count := 0
	resolved := 0
	notFound := 0
	for _, alarm := range p.Alarms {
		count++

		_, err := ns.GetComponentInfo(alarm.Alias)
		if err != nil {
			slog.Warn("Alias not found", "alias", alarm.Alias, "type", alarm.AlarmType)
			notFound++
			continue
		}
		slog.Info("Alias found", "alias", alarm.Alias, "type", alarm.AlarmType)
		resolved++
	}

	fmt.Printf("Resolved %d of %d alarms\n", resolved, count)
	fmt.Printf("Not found %d of %d alarms\n", notFound, count)

	return nil
}
