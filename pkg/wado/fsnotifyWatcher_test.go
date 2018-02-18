package wado

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mktange/wado/internal/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Watch(t *testing.T) {
	// Create tmp dir and files to watch
	tmpDir, err := ioutil.TempDir("", "wado-")
	require.NoError(t, err)

	aGoFile, err := os.Create(filepath.Join(tmpDir, "a.go"))
	require.NoError(t, err)

	aTxtFile, err := os.Create(filepath.Join(tmpDir, "a.txt"))
	require.NoError(t, err)

	glob := filepath.Join(tmpDir, "**", "*.go")

	// Setup watcher
	watcher, err := NewFsNotifyWatcher([]string{glob}, []string{})
	require.NoError(t, err)

	changeChan := watcher.CreateChangeChannel()
	assert.Len(t, changeChan, 0)

	// Make a change to a watched file
	_, err = aGoFile.WriteString("Foo")
	require.NoError(t, err)
	err = aGoFile.Sync()
	require.NoError(t, err)

	changeHappened := util.WaitForMessage(t, changeChan)
	assert.True(t, changeHappened, "Change did not appear on channel")
	assert.Len(t, changeChan, 0)

	// Make a change to a non-watched file
	_, err = aTxtFile.WriteString("Bar")
	require.NoError(t, err)
	err = aTxtFile.Sync()
	require.NoError(t, err)

	changeHappened = util.WaitForMessage(t, changeChan)
	assert.False(t, changeHappened, "Change should not appear on channel")
	assert.Len(t, changeChan, 0)
}

func Test_NewSubDir(t *testing.T) {
	// Create tmp dir and files to watch
	tmpDir, err := ioutil.TempDir("", "wado-")
	require.NoError(t, err)

	subDir := filepath.Join(tmpDir, "sub")
	err = os.Mkdir(subDir, os.ModePerm)
	require.NoError(t, err)

	aGoFile, err := os.Create(filepath.Join(subDir, "a.go"))
	require.NoError(t, err)

	glob := filepath.Join(tmpDir, "**", "*.go")

	// Setup watcher
	watcher, err := NewFsNotifyWatcher([]string{glob}, []string{})
	require.NoError(t, err)

	changeChan := watcher.CreateChangeChannel()
	assert.Len(t, changeChan, 0)

	// Make a change to a watched file
	_, err = aGoFile.WriteString("Foo")
	require.NoError(t, err)
	err = aGoFile.Sync()
	require.NoError(t, err)

	changeHappened := util.WaitForMessage(t, changeChan)
	assert.True(t, changeHappened, "Change did not appear on channel")
	assert.Len(t, changeChan, 0)
}
