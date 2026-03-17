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
