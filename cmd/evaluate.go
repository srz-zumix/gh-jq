package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/cli/go-gh/v2/pkg/jq"
)

// slurpInput reads all JSON values from the reader and returns a reader
// containing a single JSON array of those values. This pre-processes the input
// for jq's --slurp mode so that jq.EvaluateFormatted receives one array.
func slurpInput(r io.Reader) (io.Reader, error) {
	var values []json.RawMessage
	decoder := json.NewDecoder(r)
	for {
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		values = append(values, raw)
	}
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// evaluateFormatted pre-processes input based on the provided flags, then
// delegates to jq.EvaluateFormatted for expression evaluation.
func evaluateFormatted(input io.Reader, output io.Writer, expr string, indent string, colorize bool, nullInput bool, slurp bool) error {
	if nullInput {
		input = strings.NewReader("null")
	} else if slurp {
		r, err := slurpInput(input)
		if err != nil {
			return err
		}
		input = r
	}
	return jq.EvaluateFormatted(input, output, expr, indent, colorize)
}
