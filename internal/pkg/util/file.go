package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// FileStats contains a information about the status of a file and it's content.
type FileStats struct {
	Hash        string
	LastModTime time.Time
}

// GetFileStats returns a FileStats struct containing information about the file and content.
func GetFileStats(fPath string) (*FileStats, error) {
	fi, err := os.Stat(fPath)
	if err != nil {
		return nil, err
	}

	hash, err := FileHash(fPath)
	if err != nil {
		return nil, err
	}

	return &FileStats{
		Hash:        hash,
		LastModTime: fi.ModTime(),
	}, nil
}

// HasChanged checks if a file has changed.
func HasChanged(fPath string, fs *FileStats) (*FileStats, bool, error) {
	fi, err := os.Stat(fPath)
	if os.IsNotExist(err) {
		return nil, true, nil
	} else if err != nil {
		return nil, true, err
	}
	if fi.ModTime().After(fs.LastModTime) {
		hash, err := FileHash(fPath)
		if err != nil {
			return nil, true, err
		}

		if hash != fs.Hash {
			return &FileStats{
				Hash:        hash,
				LastModTime: fi.ModTime(),
			}, true, nil
		}
	}

	return nil, false, nil
}

// FileHash returns the SHA256 hash of a file's content
func FileHash(fPath string) (string, error) {
	f, err := os.Open(fPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

var replaceMap map[rune]string
var once sync.Once

func getReplaceMap() map[rune]string {
	once.Do(func() {
		replaceMap = make(map[rune]string)
		replaceMap['.'] = "\\."
		replaceMap['('] = "\\("
		replaceMap[')'] = "\\)"
		replaceMap['*'] = "[^/\\\\]+"
	})
	return replaceMap
}

// GetCouldDirMatchRegex takes a glob and returns a regex, that can be used
// with the CouldDirMatch-method, to answer if a given directory could
// eventually fullfill the glob.
func GetCouldDirMatchRegex(glob string) *regexp.Regexp {
	abs, err := filepath.Abs(glob)
	if err != nil {
		panic(err)
	}

	regexStr := filepath.ToSlash(filepath.Dir(abs))
	doubleStar := strings.Index(regexStr, "**")
	var suffix bytes.Buffer
	if doubleStar >= 0 {
		regexStr = regexStr[:doubleStar]
		_, err := suffix.WriteString(".*")
		if err != nil {
			panic(err)
		}
	}

	eMap := getReplaceMap()
	var buffer bytes.Buffer
	for _, c := range regexStr {
		if c == '/' {
			_, err := buffer.WriteString("(/")
			if err != nil {
				panic(err)
			}
			_, err = suffix.WriteString(")?")
			if err != nil {
				panic(err)
			}
		} else if rep, ok := eMap[c]; ok == true {
			_, err := buffer.WriteString(rep)
			if err != nil {
				panic(err)
			}
		} else {
			_, err := buffer.WriteRune(c)
			if err != nil {
				panic(err)
			}
		}
	}

	finalRegexStr := fmt.Sprintf("^%v%v$", buffer.String(), suffix.String())
	regex, err := regexp.Compile(finalRegexStr)
	if err != nil {
		panic(err)
	}
	return regex
}

// CouldDirMatch returns true if the given dir matches the given regex.
func CouldDirMatch(regex *regexp.Regexp, dirPath string) (bool, error) {
	abs, err := filepath.Abs(dirPath)
	if err != nil {
		return false, err
	}
	return regex.MatchString(filepath.ToSlash(abs)), nil
}

// GetLowestDirToWatch returns the lowest folder to start watching to be able
// to match all files in the given glob.
func GetLowestDirToWatch(glob string) string {
	idx := strings.IndexRune(glob, '*')
	if idx == -1 {
		return filepath.ToSlash(filepath.Dir(glob))
	}
	return filepath.ToSlash(filepath.Dir(glob[:idx]))
}
