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
	saveChannel  string
	loadChannel  string
	clearChannel string
	listFlag     bool

	nullInputFlag bool
	slurpFlag     bool
)

var rootCmd = &cobra.Command{
	Use:     "gh-jq [expression] [file...]",
	Short:   "A jq wrapper for GitHub CLI",
	Long:    `gh-jq is a jq wrapper for GitHub CLI that processes JSON using jq expressions.`,
	Version: version.Version,
	Args: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("clear") || listFlag {
			return nil
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if listFlag {
			names, err := store.List()
			if err != nil {
				return err
			}
			for _, name := range names {
				fmt.Fprintln(cmd.OutOrStdout(), name)
			}
			return nil
		}
		if cmd.Flags().Changed("clear") {
			return store.Clear(clearChannel)
		}

		expr := args[0]
		files := args[1:]

		saving := cmd.Flags().Changed("save")
		loading := cmd.Flags().Changed("load")
		if saving {
			if err := store.ValidateChannel(saveChannel); err != nil {
				return err
			}
		}

		var input io.Reader
		if nullInputFlag {
			// --null-input: do not read any input; evaluate.EvaluateFormatted will use null
		} else if loading {
			data, err := store.Load(loadChannel)
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

		if saving {
			data, err := io.ReadAll(input)
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			if err := store.Save(saveChannel, data); err != nil {
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
	f := rootCmd.Flags()
	f.StringVarP(&saveChannel, "save", "S", "", "Save input JSON to a channel for later use with --load (default channel if omitted)")
	f.StringVarP(&loadChannel, "load", "L", "", "Load previously saved JSON input from a channel (default channel if omitted)")
	f.StringVarP(&clearChannel, "clear", "C", "", "Clear saved JSON input of a channel; clears all channels if omitted")
	f.BoolVar(&listFlag, "list", false, "List channels that have saved JSON input")
	// Bare flags select a default value instead of consuming the next argument.
	f.Lookup("save").NoOptDefVal = store.DefaultChannel
	f.Lookup("load").NoOptDefVal = store.DefaultChannel
	f.Lookup("clear").NoOptDefVal = store.AllChannels
	rootCmd.MarkFlagsMutuallyExclusive("save", "load", "clear", "list")

	f.BoolVarP(&nullInputFlag, "null-input", "n", false, "Use null as the single input value")
	f.BoolVarP(&slurpFlag, "slurp", "s", false, "Read all inputs into an array")
	rootCmd.MarkFlagsMutuallyExclusive("save", "null-input")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
