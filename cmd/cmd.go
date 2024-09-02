package cmd

import (
	"go-file-server/cmd/api"
	"go-file-server/cmd/version"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "fileserver",
	Short:        "fileserver",
	SilenceUsage: true,
	Long:         `fileserver`,
	Run: func(cmd *cobra.Command, args []string) {
		tip()
	},
}

func tip() {

}

func init() {
	rootCmd.AddCommand(api.NewApiCommand())
	rootCmd.AddCommand(version.NewVersionCommand())

}

// Execute : apply commands
func Start() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
