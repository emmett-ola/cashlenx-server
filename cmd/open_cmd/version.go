package open_cmd

import (
	"fmt"
	"runtime"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/spf13/cobra"
)

var (
	Version   = model.Version
	BuildTime = "unknown"
	GitCommit = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version, build time, and git commit information`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("CashLenX v%s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
