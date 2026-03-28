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
	"github.com/srz-zumix/gh-jq/store"
	"github.com/srz-zumix/gh-jq/version"
)

var (
	saveFlag  bool
	loadFlag  bool
	clearFlag bool

	nullInputFlag      bool
	rawInputFlag       bool
	slurpFlag          bool
	argFlag            map[string]string
	argjsonFlag        map[string]string
	slurpfileFlag      map[string]string
	rawfileFlag        map[string]string
	argsFlag           []string
	jsonargsFlag       []string
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

		opts := &evalOptions{
			nullInput:          nullInputFlag,
			rawInput:           rawInputFlag,
			slurp:              slurpFlag,
			args:               argFlag,
			argjson:            argjsonFlag,
			slurpfile:          slurpfileFlag,
			rawfile:            rawfileFlag,
			positionalArgs:     argsFlag,
			positionalJSONArgs: jsonargsFlag,
		}
		return evaluateFormatted(input, out, expr, indent, colorize, opts)
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&saveFlag, "save", "S", false, "Save input JSON for later use with --load")
	rootCmd.Flags().BoolVarP(&loadFlag, "load", "L", false, "Load previously saved JSON input")
	rootCmd.Flags().BoolVarP(&clearFlag, "clear", "C", false, "Clear previously saved JSON input")
	rootCmd.MarkFlagsMutuallyExclusive("save", "load", "clear")

	rootCmd.Flags().BoolVarP(&nullInputFlag, "null-input", "n", false, "Use null as the single input value")
	rootCmd.Flags().BoolVarP(&rawInputFlag, "raw-input", "R", false, "Read each line of input as a string")
	rootCmd.Flags().BoolVarP(&slurpFlag, "slurp", "s", false, "Read all inputs into an array")
	rootCmd.Flags().StringToStringVar(&argFlag, "arg", nil, "Set variable $name to string value (format: name=value)")
	rootCmd.Flags().StringToStringVar(&argjsonFlag, "argjson", nil, "Set variable $name to JSON value (format: name=value)")
	rootCmd.Flags().StringToStringVar(&slurpfileFlag, "slurpfile", nil, "Set $name to array of JSON values from file (format: name=file)")
	rootCmd.Flags().StringToStringVar(&rawfileFlag, "rawfile", nil, "Set $name to string contents of file (format: name=file)")
	rootCmd.Flags().StringArrayVar(&argsFlag, "args", nil, "Append positional string argument to $ARGS.positional")
	rootCmd.Flags().StringArrayVar(&jsonargsFlag, "jsonargs", nil, "Append positional JSON argument to $ARGS.positional")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
