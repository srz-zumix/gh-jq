package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runRoot(args []string, stdin string) (string, error) {
	var outBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)
	rootCmd.SetIn(strings.NewReader(stdin))
	rootCmd.SetArgs(args)
	saveFlag = false
	loadFlag = false
	clearFlag = false
	nullInputFlag = false
	rawInputFlag = false
	slurpFlag = false
	argFlag = map[string]string{}
	argjsonFlag = map[string]string{}
	slurpfileFlag = map[string]string{}
	rawfileFlag = map[string]string{}
	argsFlag = nil
	jsonargsFlag = nil
	for _, name := range []string{"save", "load", "clear", "null-input", "raw-input", "slurp", "arg", "argjson", "slurpfile", "rawfile", "args", "jsonargs"} {
		rootCmd.Flags().Lookup(name).Changed = false
	}
	err := rootCmd.Execute()
	return outBuf.String(), err
}

func TestRootCmd_StringField(t *testing.T) {
	out, err := runRoot([]string{".name"}, `{"name":"test","value":42}`)
	require.NoError(t, err)
	assert.Equal(t, "test\n", out)
}

func TestRootCmd_NumberField(t *testing.T) {
	out, err := runRoot([]string{".value"}, `{"name":"test","value":42}`)
	require.NoError(t, err)
	assert.Equal(t, "42\n", out)
}

func TestRootCmd_Identity(t *testing.T) {
	out, err := runRoot([]string{"."}, `{"name":"test"}`)
	require.NoError(t, err)
	assert.Equal(t, `{"name":"test"}`+"\n", out)
}

func TestRootCmd_InvalidExpression(t *testing.T) {
	_, err := runRoot([]string{"invalid["}, `{"name":"test"}`)
	assert.Error(t, err)
}

func TestRootCmd_NoArgs(t *testing.T) {
	_, err := runRoot([]string{}, `{}`)
	assert.Error(t, err)
}

func TestRootCmd_FileInput(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "gh-jq-test-*.json")
	require.NoError(t, err)
	_, err = tmp.WriteString(`{"name":"fromfile","value":99}`)
	require.NoError(t, err)
	require.NoError(t, tmp.Close())

	out, err := runRoot([]string{".name", tmp.Name()}, "")
	require.NoError(t, err)
	assert.Equal(t, "fromfile\n", out)
}

func TestRootCmd_FileNotFound(t *testing.T) {
	_, err := runRoot([]string{".", "/nonexistent/path/file.json"}, "")
	assert.Error(t, err)
}

func TestRootCmd_SaveAndLoad(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Save input
	out, err := runRoot([]string{"--save", ".name"}, `{"name":"saved","value":1}`)
	require.NoError(t, err)
	assert.Equal(t, "saved\n", out)

	// Load saved input with different expression
	out, err = runRoot([]string{"--load", ".value"}, "")
	require.NoError(t, err)
	assert.Equal(t, "1\n", out)
}

func TestRootCmd_LoadWithoutSave(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := runRoot([]string{"--load", "."}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no saved input found")
}

func TestRootCmd_SaveOverwrite(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := runRoot([]string{"--save", "."}, `{"first":true}`)
	require.NoError(t, err)

	_, err = runRoot([]string{"--save", "."}, `{"second":true}`)
	require.NoError(t, err)

	out, err := runRoot([]string{"--load", "."}, "")
	require.NoError(t, err)
	assert.Equal(t, `{"second":true}`+"\n", out)
}

func TestRootCmd_SaveAndLoadMutuallyExclusive(t *testing.T) {
	_, err := runRoot([]string{"--save", "--load", "."}, `{}`)
	assert.Error(t, err)
}

func TestRootCmd_Clear(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Save input first
	_, err := runRoot([]string{"--save", "."}, `{"name":"saved"}`)
	require.NoError(t, err)

	// Clear saved input
	_, err = runRoot([]string{"--clear"}, "")
	require.NoError(t, err)

	// Load should now fail
	_, err = runRoot([]string{"--load", "."}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no saved input found")
}

func TestRootCmd_ClearWithoutSave(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Clear when nothing is saved should succeed
	_, err := runRoot([]string{"--clear"}, "")
	assert.NoError(t, err)
}

func TestRootCmd_ClearAndSaveMutuallyExclusive(t *testing.T) {
	_, err := runRoot([]string{"--clear", "--save", "."}, `{}`)
	assert.Error(t, err)
}

func TestRootCmd_ClearAndLoadMutuallyExclusive(t *testing.T) {
	_, err := runRoot([]string{"--clear", "--load", "."}, `{}`)
	assert.Error(t, err)
}

func TestRootCmd_NullInput(t *testing.T) {
	out, err := runRoot([]string{"--null-input", "1+1"}, "")
	require.NoError(t, err)
	assert.Equal(t, "2\n", out)
}

func TestRootCmd_NullInputExpression(t *testing.T) {
	out, err := runRoot([]string{"-n", "{a:1,b:2}"}, "")
	require.NoError(t, err)
	assert.Equal(t, `{"a":1,"b":2}`+"\n", out)
}

func TestRootCmd_RawInput(t *testing.T) {
	out, err := runRoot([]string{"--raw-input", "."}, "hello\nworld\n")
	require.NoError(t, err)
	assert.Equal(t, "hello\nworld\n", out)
}

func TestRootCmd_RawInputSlurp(t *testing.T) {
	out, err := runRoot([]string{"--raw-input", "--slurp", "."}, "hello\nworld\n")
	require.NoError(t, err)
	assert.Equal(t, "hello\nworld\n\n", out)
}

func TestRootCmd_Slurp(t *testing.T) {
	out, err := runRoot([]string{"--slurp", "length"}, `{"a":1}`+"\n"+`{"b":2}`)
	require.NoError(t, err)
	assert.Equal(t, "2\n", out)
}

func TestRootCmd_SlurpArray(t *testing.T) {
	out, err := runRoot([]string{"-s", "."}, `1`+"\n"+`2`+"\n"+`3`)
	require.NoError(t, err)
	assert.Equal(t, "[1,2,3]\n", out)
}

func TestRootCmd_Arg(t *testing.T) {
	out, err := runRoot([]string{"--arg", "name=world", "$name"}, `null`)
	require.NoError(t, err)
	assert.Equal(t, "world\n", out)
}

func TestRootCmd_Argjson(t *testing.T) {
	out, err := runRoot([]string{"--argjson", "val=42", "$val + 1"}, `null`)
	require.NoError(t, err)
	assert.Equal(t, "43\n", out)
}

func TestRootCmd_ArgjsonInvalidJSON(t *testing.T) {
	_, err := runRoot([]string{"--argjson", "val=notjson", "$val"}, `null`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--argjson val")
}

func TestRootCmd_SlurpFile(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "gh-jq-test-*.json")
	require.NoError(t, err)
	_, err = tmp.WriteString(`{"x":1}` + "\n" + `{"x":2}`)
	require.NoError(t, err)
	require.NoError(t, tmp.Close())

	out, err := runRoot([]string{"--slurpfile", "data=" + tmp.Name(), "-n", "$data | length"}, "")
	require.NoError(t, err)
	assert.Equal(t, "2\n", out)
}

func TestRootCmd_RawFile(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "gh-jq-test-*.txt")
	require.NoError(t, err)
	_, err = tmp.WriteString("hello world")
	require.NoError(t, err)
	require.NoError(t, tmp.Close())

	out, err := runRoot([]string{"--rawfile", "content=" + tmp.Name(), "-n", "$content"}, "")
	require.NoError(t, err)
	assert.Equal(t, "hello world\n", out)
}

func TestRootCmd_Args(t *testing.T) {
	out, err := runRoot([]string{"--null-input", "--args", "a", "--args", "b", "$ARGS.positional"}, "")
	require.NoError(t, err)
	assert.Equal(t, `["a","b"]`+"\n", out)
}

func TestRootCmd_Jsonargs(t *testing.T) {
	out, err := runRoot([]string{"--null-input", "--jsonargs", "1", "--jsonargs", "true", "$ARGS.positional"}, "")
	require.NoError(t, err)
	assert.Equal(t, `[1,true]`+"\n", out)
}

func TestRootCmd_ArgsNamed(t *testing.T) {
	out, err := runRoot([]string{"--null-input", "--arg", "foo=bar", "$ARGS.named"}, "")
	require.NoError(t, err)
	assert.Equal(t, `{"foo":"bar"}`+"\n", out)
}

func TestRootCmd_ArgsEmpty(t *testing.T) {
	out, err := runRoot([]string{"--null-input", "$ARGS"}, "")
	require.NoError(t, err)
	assert.Equal(t, `{"named":{},"positional":[]}`+"\n", out)
}

