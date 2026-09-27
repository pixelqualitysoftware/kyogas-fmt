package cmd

import (
	"github.com/khytryy/kyogas-fmt/formatter"
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:   "format [file]",
	Short: "Format a kyogas file",

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		formatter.Format(input)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(formatCmd)
}
