package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func Test_FileHash(t *testing.T) {

	// Create tmp dir and files to watch
	tmpDir, err := ioutil.TempDir("", "wado-")
	require.NoError(t, err)

	filePath := filepath.Join(tmpDir, "a.go")
	aFile, err := os.Create(filePath)
	require.NoError(t, err)

	_, err = aFile.WriteString("abc")
	require.NoError(t, err)

	hash, err := FileHash(filePath)
	require.NoError(t, err)

	err = aFile.Truncate(0)
	require.NoError(t, err)

	_, err = aFile.Seek(0, 0)
	require.NoError(t, err)

	_, err = aFile.WriteString("abc")
	require.NoError(t, err)

	fileContent, err := ioutil.ReadFile(filePath)
	require.NoError(t, err)

	assert.Equal(t, "abc", string(fileContent))

	newHash, err := FileHash(filePath)
	require.NoError(t, err)
	assert.Equal(t, hash, newHash)
}

func Test_CouldDirMatch_SingleStar(t *testing.T) {
	reg := GetCouldDirMatchRegex("path\\to\\my\\*\\asd.go")

	match, err := CouldDirMatch(reg, "path\\to\\my")
	require.NoError(t, err)
	assert.Equal(t, true, match)

	match, err = CouldDirMatch(reg, "path\\to\\my\\folder")
	require.NoError(t, err)
	assert.Equal(t, true, match)

	match, err = CouldDirMatch(reg, "path\\to\\my\\folder\\woop")
	require.NoError(t, err)
	assert.Equal(t, false, match)
}

func Test_CouldDirMatch_FSlash_SingleStar(t *testing.T) {
	reg := GetCouldDirMatchRegex("path/to/my/*/asd.go")

	match, err := CouldDirMatch(reg, "path/to/my")
	require.NoError(t, err)
	assert.Equal(t, true, match)

	match, err = CouldDirMatch(reg, "path/to/my/folder")
	require.NoError(t, err)
	assert.Equal(t, true, match)

	match, err = CouldDirMatch(reg, "path/to/my/folder/woop")
	require.NoError(t, err)
	assert.Equal(t, false, match)
}

func Test_CouldDirMatch_Multiple(t *testing.T) {
	reg := GetCouldDirMatchRegex("path\\*\\my\\*\\asd.go")

	match, err := CouldDirMatch(reg, "path\\to\\my\\folder")
	require.NoError(t, err)
	assert.Equal(t, true, match)

	match, err = CouldDirMatch(reg, "path\\alsoto\\my\\folder")
	require.NoError(t, err)
	assert.Equal(t, true, match)

	match, err = CouldDirMatch(reg, "path\\to\\another\\my\\folder")
	require.NoError(t, err)
	assert.Equal(t, false, match)

	match, err = CouldDirMatch(reg, "path\\to\\my\\folder\\not")
	require.NoError(t, err)
	assert.Equal(t, false, match)
}

func Test_CouldDirMatch_DoubleStar(t *testing.T) {
	reg := GetCouldDirMatchRegex("path\\to\\valid\\**\\*\\sub\\file.go")

	match, err := CouldDirMatch(reg, "path\\to\\valid\\folder")
	require.NoError(t, err)
	assert.Equal(t, true, match)

	match, err = CouldDirMatch(reg, "path\\to\\valid\\folder\\no\\matcher\\how\\itlooks\\sub\\with\\more")
	require.NoError(t, err)
	assert.Equal(t, true, match)

	match, err = CouldDirMatch(reg, "path\\to\\invalid\\folder")
	require.NoError(t, err)
	assert.Equal(t, false, match)
}

func Test_GetLowestDirToWatch(t *testing.T) {
	assert.Equal(t, "path/to/my", GetLowestDirToWatch("path/to/my/*_files.go"))
	assert.Equal(t, "path/to/my", GetLowestDirToWatch("path/to/my/**/*_files.go"))
	assert.Equal(t, "path/to/my", GetLowestDirToWatch("path/to/my/cool_*.go"))
	assert.Equal(t, "path/to/my", GetLowestDirToWatch("path/to/my/cool_file.go"))
	assert.Equal(t, "path/to/my", GetLowestDirToWatch("path/to/my/**/*/very/deep/*_files.go"))
	assert.Equal(t, "../..", GetLowestDirToWatch("../../**/*/sub/dir/*_files.go"))
	assert.Equal(t, ".", GetLowestDirToWatch("."))
}
