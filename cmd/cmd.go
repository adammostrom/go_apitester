package cmd

import (
	"fmt"
	"main/models"
	"main/runner"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "apitest", // To be decided
	Short: "YAML based api testing tool",
	//Long:  "Run, validate and compare API test scenarios defined in YAML files.",
}

// The cobra.Command is basically a description of a CLI command.
/*
"Create a command called run, which expects one argument, has this description, and execute this function."
*/

var runCmd = &cobra.Command{
	Use:   "run <file>",
	Short: "Run API tests",

	Args: cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]

		config := models.RunConfig{
			Repeat:   repeat,
			Verbose:  verbose,
			Output:   output,
			Requests: requests,
		}
		results, err := runner.RunTests(config, file)
		if err != nil {
			return err
		}

		fmt.Println(results)

		return nil
	},
}

/*
Structure:

Core Commands:
	run -flags
	validate <file.yaml>
	generate [args]
	compare
	single -flags (generates basic yaml and then sends it)
	validate tests (correct/analyze/inspect .csv file)
	version
	repeating


                    rootCmd
                       │
              ┌────────┼────────┐
              │        │        │
             run    validate   version
              │
        ┌─────┼──────┐
        │     │      │
       args  flags  Run()

*/

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {

	// Nested commands
	rootCmd.AddCommand(runCmd)
	/* 	rootCmd.AddCommand(configCmd)

	   	configCmd.AddCommand(configSetCmd)
	   	configCmd.AddCommand(configShowCmd) */

	runCmd.Flags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false, // default value
		"Enable verbose output",
	)

	runCmd.Flags().IntVarP(
		&parallel,
		"parallel",
		"p",
		1,
		"Number of threads",
	)

	runCmd.Flags().StringVarP(
		&output,
		"output",
		"o",
		"",
		"Output file",
	)

	runCmd.Flags().IntVarP(
		&requests,
		"requests",
		"n",
		1,
		"Number of requests",
	)

	runCmd.Flags().IntVarP(
		&repeat,
		"repeats",
		"r",
		0,
		"number of repeats for the specified requests",
	)
}

// FLAGS

var (
	verbose  bool
	parallel int // Amount of threads
	output   string
	requests int
	repeat   int
)

// Examples

/*

apitest run tests.yaml --verbose
apitest run tests.yaml -v

*/
