package psalerts

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/3ideas/psasim/lib/namerif"
)

func CheckAliases(nameChecker namerif.NameService) {
	// Iteratively read a alias from stdin and check it
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter alias: ")
		alias, _ := reader.ReadString('\n')
		alias = strings.TrimSpace(alias)
		_, err := nameChecker.GetComponentInfo(alias)
		if err != nil {
			fmt.Println("Alias not found")
		} else {
			fmt.Println("Alias found")
		}
	}
}
