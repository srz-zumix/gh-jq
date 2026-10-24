package evaluate

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateFormatted(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		input     string
		nullInput bool
		slurp     bool
		want      string
	}{
		{name: "string field", expr: ".name", input: `{"name":"test","value":42}`, want: "test\n"},
		{name: "number field", expr: ".value", input: `{"name":"test","value":42}`, want: "42\n"},
		{name: "identity", expr: ".", input: `{"name":"test"}`, want: `{"name":"test"}` + "\n"},
		{name: "null input arithmetic", expr: "1+1", nullInput: true, want: "2\n"},
		{name: "null input object", expr: "{a:1,b:2}", nullInput: true, want: `{"a":1,"b":2}` + "\n"},
		{name: "null input identity", expr: ".", nullInput: true, want: "\n"},
		{name: "null input ignores input", expr: ".", input: `{"name":"test"}`, nullInput: true, want: "\n"},
		{name: "slurp length", expr: "length", input: `{"a":1}` + "\n" + `{"b":2}`, slurp: true, want: "2\n"},
		{name: "slurp array", expr: ".", input: "1\n2\n3", slurp: true, want: "[1,2,3]\n"},
		{name: "slurp single", expr: ".", input: `{"x":1}`, slurp: true, want: `[{"x":1}]` + "\n"},
		{name: "null input takes precedence over slurp", expr: ".", nullInput: true, slurp: true, want: "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := EvaluateFormatted(strings.NewReader(tt.input), &out, tt.expr, "", false, tt.nullInput, tt.slurp)
			require.NoError(t, err)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestEvaluateFormatted_InvalidExpression(t *testing.T) {
	var out bytes.Buffer
	err := EvaluateFormatted(strings.NewReader(`{"name":"test"}`), &out, "invalid[", "", false, false, false)
	assert.Error(t, err)
}

func TestEvaluateFormatted_SlurpInvalidJSON(t *testing.T) {
	var out bytes.Buffer
	err := EvaluateFormatted(strings.NewReader(`{"a":`), &out, ".", "", false, false, true)
	assert.Error(t, err)
}
