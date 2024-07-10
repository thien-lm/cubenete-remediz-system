package manipulation

import (
	"fmt"
	"strings"
)

func getVAppNameFromVMName(vmName string) string {
	// Split the string on "-worker" and take the first part
	parts := strings.Split(vmName, "-worker")
	if len(parts) > 0 {
		 // This will print "fke-mps-ko-xoa-2i53tqyp"
	} else {
		fmt.Println("The specified pattern was not found in the string.")
	}
	return parts[0]
}