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
	rootCmd.Flags().Lookup("save").Changed = false
	rootCmd.Flags().Lookup("load").Changed = false
	rootCmd.Flags().Lookup("clear").Changed = false
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
