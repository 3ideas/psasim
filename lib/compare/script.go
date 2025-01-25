package compare

import (
	"fmt"
	"os"
	"sort"
)

func GenerateNameQueryScript(alarmComparison *AlarmsComparison, filename string) error {

	// get the unique aliases
	uniqueAliases := make(map[string]struct{})
	for _, alarm := range alarmComparison.alarms {
		uniqueAliases[alarm.POAlias] = struct{}{}
	}

	// write the script
	// sort the aliases
	aliases := make([]string, 0)
	for alias := range uniqueAliases {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)

	// Open the file
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the script
	for _, alias := range aliases {
		if alias == "" || alias == "/" {
			continue
		}
		fmt.Fprintf(f, "echo %s\n", alias)
		fmt.Fprintf(f, "get_comp_desc %s\n", alias)
	}

	return nil

}
