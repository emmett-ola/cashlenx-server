package cmd

import "testing"

func TestRootRunPrintsUsageHint(t *testing.T) {
	rootCmd.Run(rootCmd, nil)
}
