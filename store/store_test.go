package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoad(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	data := []byte(`{"name":"test","value":42}`)
	err := Save(data)
	require.NoError(t, err)

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, data, loaded)
}

func TestLoad_NoSavedData(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no saved input found")
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	data := []byte(`{"key":"value"}`)
	err := Save(data)
	require.NoError(t, err)

	path := filepath.Join(tmpDir, "gh-jq", defaultFileName)
	_, err = os.Stat(path)
	assert.NoError(t, err)
}

func TestSave_OverwritesExisting(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	err := Save([]byte(`{"first":true}`))
	require.NoError(t, err)

	err = Save([]byte(`{"second":true}`))
	require.NoError(t, err)

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"second":true}`), loaded)
}

func TestClear(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	err := Save([]byte(`{"name":"test"}`))
	require.NoError(t, err)

	err = Clear()
	require.NoError(t, err)

	_, err = Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no saved input found")
}

func TestClear_NoSavedData(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	err := Clear()
	assert.NoError(t, err)
}
