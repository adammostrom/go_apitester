package printout

import "fmt"

var Verbose bool

// Printer for verbose messages if verbose flag is true
func VPrintf(format string, args ...any) {
	if Verbose {
		fmt.Printf(format, args...)
	}
}
