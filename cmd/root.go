/*
Copyright © 2025 srz_zumix
*/
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/cli/go-gh/v2/pkg/jq"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-jq/store"
	"github.com/srz-zumix/gh-jq/version"
)

var (
	saveFlag bool
	loadFlag bool
)

var rootCmd = &cobra.Command{
	Use:     "gh-jq [expression] [file...]",
	Short:   "A jq wrapper for GitHub CLI",
	Long:    `gh-jq is a jq wrapper for GitHub CLI that processes JSON using jq expressions.`,
	Version: version.Version,
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		expr := args[0]
		files := args[1:]

		var input io.Reader
		if loadFlag {
			data, err := store.Load()
			if err != nil {
				return err
			}
			input = bytes.NewReader(data)
		} else if len(files) == 0 {
			input = cmd.InOrStdin()
		} else {
			opened := make([]*os.File, 0, len(files))
			defer func() {
				for _, f := range opened {
					f.Close() // nolint:errcheck
				}
			}()
			readers := make([]io.Reader, 0, len(files))
			for _, name := range files {
				f, err := os.Open(name)
				if err != nil {
					return fmt.Errorf("cannot open file %q: %w", name, err)
				}
				opened = append(opened, f)
				readers = append(readers, f)
			}
			input = io.MultiReader(readers...)
		}

		if saveFlag {
			data, err := io.ReadAll(input)
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			if err := store.Save(data); err != nil {
				return err
			}
			input = bytes.NewReader(data)
		}

		out := cmd.OutOrStdout()
		isTTY := isatty.IsTerminal(os.Stdout.Fd())
		indent := ""
		if isTTY {
			indent = "  "
		}
		colorize := isTTY

		return jq.EvaluateFormatted(input, out, expr, indent, colorize)
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&saveFlag, "save", "S", false, "Save input JSON for later use with --load")
	rootCmd.Flags().BoolVarP(&loadFlag, "load", "L", false, "Load previously saved JSON input")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
