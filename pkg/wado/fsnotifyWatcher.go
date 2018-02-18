package wado

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/fsnotify/fsnotify"
	zglob "github.com/mattn/go-zglob"
	"github.com/mattn/go-zglob/fastwalk"
	"github.com/mktange/wado/internal/pkg/syncimpls"
	"github.com/mktange/wado/internal/pkg/util"
)

type fsnotifyWatcher struct {
	includeGlobs    []string
	excludeGlobs    []string
	includeDirRegex []*regexp.Regexp
	changeChans     []chan string

	watchedFiles  *syncimpls.MapStringFileStats
	watchedDirs   *syncimpls.MapStringFsWatch
	callbackFuncs []func(string)
	callbackLock  *sync.Mutex
}

func (watcher *fsnotifyWatcher) FileCount() int {
	return watcher.watchedFiles.Size()
}

func (watcher *fsnotifyWatcher) CreateChangeChannel() chan string {
	newChan := make(chan string, 20)
	watcher.changeChans = append(watcher.changeChans, newChan)
	return newChan
}

func (watcher *fsnotifyWatcher) Close() error {
	watcher.watchedFiles.Clear()
	return watcher.watchedDirs.ClearAndClose()
}

func (watcher *fsnotifyWatcher) AddCallback(cb func(string)) {
	watcher.callbackLock.Lock()
	watcher.callbackFuncs = append(watcher.callbackFuncs, cb)
	watcher.callbackLock.Unlock()
}

// NewFsNotifyWatcher creates a new watcher based on the given configurations using fsnotify.
// Changes are posted on the channel that can be gotten with GetChannel()
func NewFsNotifyWatcher(includeGlobs []string, excludeGlobs []string) (Watcher, error) {
	includeDirRegex := []*regexp.Regexp{}
	for _, glob := range includeGlobs {
		includeDirRegex = append(includeDirRegex, util.GetCouldDirMatchRegex(glob))
	}

	watcher := &fsnotifyWatcher{
		includeGlobs:    includeGlobs,
		excludeGlobs:    excludeGlobs,
		includeDirRegex: includeDirRegex,
		changeChans:     []chan string{},

		watchedFiles:  syncimpls.NewMapStringFileStats(),
		watchedDirs:   syncimpls.NewMapStringFsWatch(),
		callbackFuncs: []func(string){},
		callbackLock:  &sync.Mutex{},
	}

	for _, glob := range includeGlobs {
		baseDir := util.GetLowestDirToWatch(glob)
		_, err := os.Stat(baseDir)
		if err != nil {
			return nil, err
		}
		fullPath, err := filepath.Abs(baseDir)
		if err != nil {
			return nil, err
		}

		watcher.walkAndWatch(fullPath)
	}

	return watcher, nil
}

func (watcher *fsnotifyWatcher) walkAndWatch(watchPath string) error {
	return fastwalk.FastWalk(watchPath, func(currentPath string, info os.FileMode) error {
		var err error
		if info.IsDir() {
			err = watcher.watchDir(currentPath)
		} else {
			watcher.watchFileIfMatch(currentPath)
		}
		return err
	})
}

func (watcher *fsnotifyWatcher) watchFileIfMatch(fPath string) {
	if watcher.shouldWatchFile(fPath) {
		fs, err := util.GetFileStats(fPath)
		if err != nil {
			log.Println("Error when checking file:", err)
		}
		watcher.watchedFiles.Store(fPath, fs)
	}
}

func (watcher *fsnotifyWatcher) shouldWatchDir(fPath string) bool {
	for _, regex := range watcher.includeDirRegex {
		couldMatch, err := util.CouldDirMatch(regex, fPath)
		if err == nil && couldMatch {
			return true
		}
	}
	return false
}

func (watcher *fsnotifyWatcher) shouldWatchFile(fPath string) bool {
	for _, fileGlob := range watcher.excludeGlobs {
		matched, err := zglob.Match(fileGlob, fPath)
		if err != nil {
			panic(err)
		}
		if matched {
			return false
		}
	}
	for _, fileGlob := range watcher.includeGlobs {
		matched, err := zglob.Match(fileGlob, fPath)
		if err != nil {
			panic(err)
		}
		if matched {
			return true
		}
	}
	return false
}

func (watcher *fsnotifyWatcher) changeDetected(filePath string) {
	for _, cb := range watcher.callbackFuncs {
		go cb(filePath)
	}

	for _, ch := range watcher.changeChans {
		select {
		case ch <- filePath:
		default:
		}
	}
}

func (watcher *fsnotifyWatcher) watchDir(watchPath string) error {
	if _, exists := watcher.watchedDirs.Load(watchPath); exists == true || !watcher.shouldWatchDir(watchPath) {
		return nil
	}

	fsWatch, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = fsWatch.Add(watchPath)
	if err != nil {
		return err
	}
	watcher.watchedDirs.Store(watchPath, fsWatch)

	// Start listening on the events channel
	go func() {
		for {
			select {
			case event := <-fsWatch.Events:
				filePath := event.Name
				fi, err := os.Stat(filePath)
				if err != nil {
					panic(err)
				}
				isDir := fi.Mode().IsDir()

				if event.Op&fsnotify.Create == fsnotify.Create { // Created
					if isDir {
						watcher.watchDir(filePath)
					} else {
						watcher.watchFileIfMatch(filePath)
					}

				} else if event.Op&fsnotify.Remove == fsnotify.Remove { // Removed
					if isDir {
						if dir, ok := watcher.watchedDirs.Load(filePath); ok == true {
							go dir.Close()
							watcher.watchedDirs.Delete(filePath)
						}

					} else {
						watcher.watchedFiles.Delete(filePath)
					}

				} else if event.Op&fsnotify.Write == fsnotify.Write { // Changed
					if !isDir {
						if fs, ok := watcher.watchedFiles.Load(filePath); ok == true {
							newFs, changed, err := util.HasChanged(filePath, fs)
							if err != nil {
								log.Println("Error:", err)
							}

							if changed {
								if newFs == nil {
									watcher.watchedFiles.Delete(filePath)
								} else {
									watcher.watchedFiles.Store(filePath, newFs)
									watcher.changeDetected(filePath)
								}
							}
						}
					}
				}

			case err := <-fsWatch.Errors:
				log.Println("Error while watching:", err)
			}
		}
	}()

	return nil
}
