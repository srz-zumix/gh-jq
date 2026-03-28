// Updated evaluate.go to prioritize go-gh package and restrict input options.

package cmd

import (
	"encoding/json"
	"io"
	"github.com/cli/go-gh/v2/pkg/jq"
)

// Restrict input options
func evaluateInput(opts *evalOptions, reader io.Reader) (io.Reader, error) {
	if opts.nullInput {
		return buildSingleInputJSON(nil), nil
	} else if opts.slurp {
		return buildSingleInputJSON(reader, opts)
	} else {
		return nil, fmt.Errorf("unsupported option: %s", opts)
	}
}

// buildSingleInputJSON creates a reader over JSON bytes based on input type.
func buildSingleInputJSON(reader io.Reader, opts *evalOptions) (io.Reader, error) {
	var input interface{}
	if opts.slurp {
		// Read all inputs into an array
		var inputs []interface{}
		decoder := json.NewDecoder(reader)
		for {
			var value interface{}
			if err := decoder.Decode(&value); err != nil {
				return nil, err
			}
			inputs = append(inputs, value)
		}
		return json.NewEncoder(inputs), nil
	}

	// Default input handling
	if opts.defaultInput {
		if err := json.NewDecoder(reader).Decode(&input); err != nil {
			return nil, err
		}
	}

	return json.NewEncoder(input), nil
}

// Main function to call jq.EvaluateFormatted
func runEval(reader io.Reader, output io.Writer, expr string, indent bool, colorize bool) error {
	inputReader, err := evaluateInput(opts, reader)
	if err != nil {
		return err
	}

	return jq.EvaluateFormatted(inputReader, output, expr, indent, colorize)
}

// Remove unnecessary functions and maintain minimal imports.
