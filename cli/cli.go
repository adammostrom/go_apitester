package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tbd", // To be decided
	Short: "YAML based api testing tool",
	Long:  "Run, validate and compare API test scenarios defined in YAML files.",
}

/*
Structure:

Core Commands:
	run -flags
	validate <file.yaml>
	generate [args]
	compare
	single -flags (generates basic yaml and then sends it)




*/

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
