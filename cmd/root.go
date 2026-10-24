/*
Copyright © 2025 srz_zumix
*/
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/srz-zumix/gh-jq/internal/evaluate"
	"github.com/srz-zumix/gh-jq/internal/store"
	"github.com/srz-zumix/gh-jq/version"
)

var (
	saveFlag  bool
	loadFlag  bool
	clearFlag bool

	nullInputFlag bool
	slurpFlag     bool
)

var rootCmd = &cobra.Command{
	Use:     "gh-jq [expression] [file...]",
	Short:   "A jq wrapper for GitHub CLI",
	Long:    `gh-jq is a jq wrapper for GitHub CLI that processes JSON using jq expressions.`,
	Version: version.Version,
	Args: func(cmd *cobra.Command, args []string) error {
		if clearFlag {
			return nil
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if clearFlag {
			return store.Clear()
		}

		expr := args[0]
		files := args[1:]

		var input io.Reader
		if nullInputFlag {
			// --null-input: do not read any input; evaluate.EvaluateFormatted will use null
		} else if loadFlag {
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

		return evaluate.EvaluateFormatted(input, out, expr, indent, colorize, nullInputFlag, slurpFlag)
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&saveFlag, "save", "S", false, "Save input JSON for later use with --load")
	rootCmd.Flags().BoolVarP(&loadFlag, "load", "L", false, "Load previously saved JSON input")
	rootCmd.Flags().BoolVarP(&clearFlag, "clear", "C", false, "Clear previously saved JSON input")
	rootCmd.MarkFlagsMutuallyExclusive("save", "load", "clear")

	rootCmd.Flags().BoolVarP(&nullInputFlag, "null-input", "n", false, "Use null as the single input value")
	rootCmd.Flags().BoolVarP(&slurpFlag, "slurp", "s", false, "Read all inputs into an array")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
