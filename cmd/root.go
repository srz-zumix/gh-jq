/*
Copyright © 2025 srz_zumix
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-jq/version"
)

var rootCmd = &cobra.Command{
	Use:     "gh-jq",
	Short:   "A jq wrapper for GitHub CLI",
	Long:    `gh-jq is a jq wrapper for GitHub CLI.`,
	Version: version.Version,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
