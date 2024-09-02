package version

import (
	"fmt"
	"go-file-server/internal/common/global"

	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Get version info",
		Example: "fileserver version",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
}

func run() {
	fmt.Println(global.Version)
}
