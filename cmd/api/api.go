package api

import (
	"go-file-server/internal/app"
	"go-file-server/pkgs/config"
	"go-file-server/pkgs/zlog"
	"path/filepath"

	"github.com/spf13/cobra"
)

const appName = "fileserver"

func NewApiCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "server",
		Short: "文件服务器",
		Long:  `go-file-server server -c ./config.yaml`,
		PreRun: func(cmd *cobra.Command, args []string) {
			setup(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.Start()
		},
	}
	rootCmd.Flags().StringP("config", "c", "./config.yaml", "use -c set Your congfile")
	return rootCmd
}

func setup(cmd *cobra.Command) {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		panic(err)
	}
	config.Init(
		config.SetFile(configFile),
		config.SetEnvPrefix(appName),
		config.SetAutomaticEnv(),
		config.Setflags(cmd.Flags()),
	)
	zlog.Init(
		zlog.WithPath(filepath.Join(config.LoggerCfg.Path, appName)),
		zlog.WithLevel(config.LoggerCfg.Level))
}
