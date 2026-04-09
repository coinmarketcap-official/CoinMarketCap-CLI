package cmd

import (
	"fmt"
	"os"

	"github.com/openCMC/CoinMarketCap-CLI/internal/display"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:               "cmc",
	Short:             "CoinMarketCap CLI for cryptocurrency market data",
	Long:              "A command-line tool for accessing CoinMarketCap cryptocurrency market data.",
	Version:           version,
	PersistentPreRunE: validateRootFlags,
	Run: func(cmd *cobra.Command, args []string) {
		display.PrintLogo()
		display.PrintWelcomeBox()
	},
}

func Execute() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	if err := rootCmd.Execute(); err != nil {
		// Emit structured JSON error to stderr when -o json, otherwise plain text.
		cmd, _, _ := rootCmd.Find(os.Args[1:])
		if cmd != nil && outputJSON(cmd) {
			_ = formatError(cmd, err)
		} else {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
		os.Exit(1)
	}
}

func addOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "json", "Output format (json, table)")
}

func validateRootFlags(cmd *cobra.Command, _ []string) error {
	return validateOutputFormat(cmd)
}

func validateOutputFormat(cmd *cobra.Command) error {
	if cmd == nil || cmd.Flags().Lookup("output") == nil {
		return nil
	}

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
	switch output {
	case "json", "table":
		return nil
	default:
		return fmt.Errorf("--output must be one of: json, table (got %q)", output)
	}
}

func outputJSON(cmd *cobra.Command) bool {
	o, err := cmd.Flags().GetString("output")
	return err == nil && o == "json"
}
