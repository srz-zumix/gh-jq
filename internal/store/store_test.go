package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setConfigDir redirects the config directory to a temporary directory for the test.
func setConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	orig := userConfigDir
	userConfigDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { userConfigDir = orig })
	return dir
}

func TestSaveAndLoad(t *testing.T) {
	setConfigDir(t)

	data := []byte(`{"name":"test","value":42}`)
	err := Save(DefaultChannel, data)
	require.NoError(t, err)

	loaded, err := Load(DefaultChannel)
	require.NoError(t, err)
	assert.Equal(t, data, loaded)
}

func TestLoad_NoSavedData(t *testing.T) {
	setConfigDir(t)

	_, err := Load(DefaultChannel)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no saved input found")
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := setConfigDir(t)

	data := []byte(`{"key":"value"}`)
	err := Save(DefaultChannel, data)
	require.NoError(t, err)

	path := filepath.Join(tmpDir, "gh-jq", defaultFileName)
	_, err = os.Stat(path)
	assert.NoError(t, err)
}

func TestSave_OverwritesExisting(t *testing.T) {
	setConfigDir(t)

	err := Save(DefaultChannel, []byte(`{"first":true}`))
	require.NoError(t, err)

	err = Save(DefaultChannel, []byte(`{"second":true}`))
	require.NoError(t, err)

	loaded, err := Load(DefaultChannel)
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"second":true}`), loaded)
}

func TestClear(t *testing.T) {
	setConfigDir(t)

	err := Save(DefaultChannel, []byte(`{"name":"test"}`))
	require.NoError(t, err)

	err = Clear(DefaultChannel)
	require.NoError(t, err)

	_, err = Load(DefaultChannel)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no saved input found")
}

func TestClear_NoSavedData(t *testing.T) {
	setConfigDir(t)

	err := Clear(DefaultChannel)
	assert.NoError(t, err)
}

func TestChannels_AreIndependent(t *testing.T) {
	setConfigDir(t)

	require.NoError(t, Save("alpha", []byte(`{"a":1}`)))
	require.NoError(t, Save("beta", []byte(`{"b":2}`)))
	require.NoError(t, Save(DefaultChannel, []byte(`{"d":3}`)))

	loaded, err := Load("alpha")
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"a":1}`), loaded)

	loaded, err = Load("beta")
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"b":2}`), loaded)

	loaded, err = Load(DefaultChannel)
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"d":3}`), loaded)
}

func TestLoad_MissingChannel(t *testing.T) {
	setConfigDir(t)

	_, err := Load("missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no saved input found for channel \"missing\"")
}

func TestClear_NamedChannelOnly(t *testing.T) {
	setConfigDir(t)

	require.NoError(t, Save("alpha", []byte(`{"a":1}`)))
	require.NoError(t, Save("beta", []byte(`{"b":2}`)))

	require.NoError(t, Clear("alpha"))

	_, err := Load("alpha")
	assert.Error(t, err)
	_, err = Load("beta")
	assert.NoError(t, err)
}

func TestClear_AllChannels(t *testing.T) {
	tmpDir := setConfigDir(t)

	require.NoError(t, Save("alpha", []byte(`{"a":1}`)))
	require.NoError(t, Save(DefaultChannel, []byte(`{"d":1}`)))
	require.NoError(t, Save("beta", []byte(`{"b":2}`)))

	// Unrelated files in the channel directory must survive.
	other := filepath.Join(tmpDir, "gh-jq", channelDirName, "notes.txt")
	require.NoError(t, os.WriteFile(other, []byte("keep"), filePerm))

	require.NoError(t, Clear(AllChannels))

	names, err := List()
	require.NoError(t, err)
	assert.Empty(t, names)
	_, err = os.Stat(other)
	assert.NoError(t, err)
}

func TestClear_AllChannels_NoSavedData(t *testing.T) {
	setConfigDir(t)

	assert.NoError(t, Clear(AllChannels))
}

func TestList(t *testing.T) {
	setConfigDir(t)

	names, err := List()
	require.NoError(t, err)
	assert.Empty(t, names)

	require.NoError(t, Save("zeta", []byte(`{}`)))
	require.NoError(t, Save("alpha", []byte(`{}`)))
	require.NoError(t, Save(DefaultChannel, []byte(`{}`)))

	names, err = List()
	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", DefaultChannel, "zeta"}, names)
}

func TestValidateChannel(t *testing.T) {
	valid := []string{"default", "alpha", "my-channel_1", "0"}
	for _, name := range valid {
		assert.NoError(t, ValidateChannel(name), name)
	}

	invalid := []string{"", "*", "Alpha", "../escape", "a/b", "a b", "a.json", "日本語"}
	for _, name := range invalid {
		assert.Error(t, ValidateChannel(name), name)
	}
}

func TestSave_RejectsInvalidChannel(t *testing.T) {
	setConfigDir(t)

	assert.Error(t, Save("../escape", []byte(`{}`)))
	assert.Error(t, Save("", []byte(`{}`)))
}
