/*
Copyright © 2025 srz_zumix
*/
package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"strconv"

	"github.com/cli/go-gh/v2/pkg/jsonpretty"
	"github.com/itchyny/gojq"
)

// evalOptions holds the jq input option flags.
type evalOptions struct {
	nullInput          bool
	rawInput           bool
	slurp              bool
	args               map[string]string
	argjson            map[string]string
	slurpfile          map[string]string
	rawfile            map[string]string
	positionalArgs     []string
	positionalJSONArgs []string
}

// evaluateFormatted evaluates a jq expression with the given options.
// It replicates the behavior of jq.EvaluateFormatted but supports additional
// input options such as --null-input, --raw-input, --slurp, --arg, --argjson,
// --slurpfile, --rawfile, --args, and --jsonargs.
func evaluateFormatted(input io.Reader, output io.Writer, expr, indent string, colorize bool, opts *evalOptions) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("failed to parse jq expression: %w", err)
	}

	varNames, varValues, err := buildVariables(opts)
	if err != nil {
		return err
	}

	code, err := gojq.Compile(
		query,
		gojq.WithEnvironLoader(os.Environ),
		gojq.WithVariables(varNames),
	)
	if err != nil {
		return err
	}

	inputValues, err := getInputValues(input, opts)
	if err != nil {
		return err
	}

	enc := &jqEncoder{w: output, indent: indent, colorize: colorize}
	for _, inputVal := range inputValues {
		iter := code.Run(inputVal, varValues...)
		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			if runErr, isErr := v.(error); isErr {
				var haltErr *gojq.HaltError
				if errors.As(runErr, &haltErr) && haltErr.Value() == nil {
					return nil
				}
				return runErr
			}
			if err := enc.write(v); err != nil {
				return err
			}
		}
	}
	return nil
}

// getInputValues reads inputs based on the given options.
func getInputValues(input io.Reader, opts *evalOptions) ([]interface{}, error) {
	if opts.nullInput {
		return []interface{}{nil}, nil
	}

	if opts.rawInput && opts.slurp {
		data, err := io.ReadAll(input)
		if err != nil {
			return nil, err
		}
		return []interface{}{string(data)}, nil
	}

	if opts.rawInput {
		var values []interface{}
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			values = append(values, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return values, nil
	}

	if opts.slurp {
		dec := json.NewDecoder(input)
		var arr []interface{}
		for {
			var v interface{}
			if err := dec.Decode(&v); err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}
		if arr == nil {
			arr = []interface{}{}
		}
		return []interface{}{arr}, nil
	}

	// Default: parse single JSON value
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return []interface{}{v}, nil
}

// buildVariables constructs variable names and values from evalOptions.
func buildVariables(opts *evalOptions) ([]string, []interface{}, error) {
	var names []string
	var values []interface{}
	named := map[string]interface{}{}

	for k, v := range opts.args {
		names = append(names, "$"+k)
		values = append(values, v)
		named[k] = v
	}

	for k, v := range opts.argjson {
		var parsed interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return nil, nil, fmt.Errorf("--argjson %s: invalid JSON: %w", k, err)
		}
		names = append(names, "$"+k)
		values = append(values, parsed)
		named[k] = parsed
	}

	for k, path := range opts.slurpfile {
		arr, err := slurpJSONFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("--slurpfile %s: %w", k, err)
		}
		names = append(names, "$"+k)
		values = append(values, arr)
		named[k] = arr
	}

	for k, path := range opts.rawfile {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("--rawfile %s: cannot read %q: %w", k, path, err)
		}
		names = append(names, "$"+k)
		values = append(values, string(data))
		named[k] = string(data)
	}

	// Build positional args
	positional := []interface{}{}
	for _, v := range opts.positionalArgs {
		positional = append(positional, v)
	}
	for _, v := range opts.positionalJSONArgs {
		var parsed interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return nil, nil, fmt.Errorf("--jsonargs: invalid JSON %q: %w", v, err)
		}
		positional = append(positional, parsed)
	}

	// $ARGS is always available
	names = append(names, "$ARGS")
	values = append(values, map[string]interface{}{
		"named":      named,
		"positional": positional,
	})

	return names, values, nil
}

// jqEncoder writes jq output values, replicating go-gh's EvaluateFormatted behavior:
// scalar values are written as raw strings; complex values are written as JSON.
type jqEncoder struct {
	w        io.Writer
	indent   string
	colorize bool
}

func (e *jqEncoder) write(v interface{}) error {
	if text, err := jqScalarToString(v); err == nil {
		_, writeErr := fmt.Fprintln(e.w, text)
		return writeErr
	}
	return e.encodeJSON(v)
}

func (e *jqEncoder) encodeJSON(v interface{}) error {
	var b []byte
	var err error
	if e.indent == "" {
		b, err = json.Marshal(v)
	} else {
		b, err = json.MarshalIndent(v, "", e.indent)
	}
	if err != nil {
		return err
	}
	if !e.colorize {
		if _, err := e.w.Write(b); err != nil {
			return err
		}
		_, err = e.w.Write([]byte{'\n'})
		return err
	}
	return jsonpretty.Format(e.w, bytes.NewReader(b), e.indent, true)
}

// jqScalarToString converts a scalar jq output value to a string.
// Returns an error for non-scalar (complex) types.
func jqScalarToString(v interface{}) (string, error) {
	switch tv := v.(type) {
	case string:
		return tv, nil
	case int:
		return strconv.Itoa(tv), nil
	case float64:
		if math.Trunc(tv) == tv {
			return strconv.FormatFloat(tv, 'f', 0, 64), nil
		}
		return strconv.FormatFloat(tv, 'f', 2, 64), nil
	case *big.Int:
		return tv.String(), nil
	case nil:
		return "", nil
	case bool:
		return fmt.Sprintf("%v", tv), nil
	default:
		return "", fmt.Errorf("not a scalar: %T", tv)
	}
}

// slurpJSONFile reads all JSON values from a file and returns them as a slice.
func slurpJSONFile(path string) ([]interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %q: %w", path, err)
	}
	defer f.Close() // nolint:errcheck
	dec := json.NewDecoder(f)
	arr := []interface{}{}
	for {
		var item interface{}
		if err := dec.Decode(&item); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("invalid JSON in %q: %w", path, err)
		}
		arr = append(arr, item)
	}
	return arr, nil
}
