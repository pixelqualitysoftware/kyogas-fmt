package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kyogas-fmt",
	Short: "Formatting tool for kyogas files",
	Long:  `kyogas-fmt is a CLI tool for formatting kyogas files.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
