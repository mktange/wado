package wado

import (
	"log"
	"sync"
	"time"

	zglob "github.com/mattn/go-zglob"
	"github.com/mktange/wado/internal/pkg/syncimpls"
	"github.com/mktange/wado/internal/pkg/util"
)

type pollWatcher struct {
	includeGlobs []string
	excludeGlobs []string
	changeChans  []chan string

	watchedFiles *syncimpls.MapStringFileStats
	done         chan bool

	callbackFuncs []func(string)
	callbackLock  *sync.Mutex
}

func (watcher *pollWatcher) FileCount() int {
	return watcher.watchedFiles.Size()
}

func (watcher *pollWatcher) CreateChangeChannel() chan string {
	newChan := make(chan string, 20)
	watcher.changeChans = append(watcher.changeChans, newChan)
	return newChan
}

func (watcher *pollWatcher) Close() error {
	watcher.watchedFiles.Clear()
	close(watcher.done)
	return nil
}

func (watcher *pollWatcher) AddCallback(cb func(string)) {
	watcher.callbackLock.Lock()
	defer watcher.callbackLock.Unlock()
	watcher.callbackFuncs = append(watcher.callbackFuncs, cb)
}

// NewPollWatcher creates a new watcher based on the given configurations using polling.
// Changes are posted on the channel that can be gotten with GetChannel()
func NewPollWatcher(includeGlobs []string, excludeGlobs []string) (Watcher, error) {
	watcher := &pollWatcher{
		includeGlobs: includeGlobs,
		excludeGlobs: excludeGlobs,
		changeChans:  []chan string{},
		done:         make(chan bool, 10),

		watchedFiles:  syncimpls.NewMapStringFileStats(),
		callbackFuncs: []func(string){},
		callbackLock:  &sync.Mutex{},
	}

	watcher.addAllFromGlobs(false)
	go watcher.checkFiles(300)
	go watcher.checkGlobs(1500)

	return watcher, nil
}

func (watcher *pollWatcher) addAllFromGlobs(triggerChange bool) {
	for _, glob := range watcher.includeGlobs {
		files, err := zglob.Glob(glob)
		if err != nil {
			log.Println("Error:", err)
			continue
		}

		for _, file := range files {
			if _, ok := watcher.watchedFiles.Load(file); ok == false {
				fs, err := util.GetFileStats(file)
				if err != nil {
					log.Println("Error:", err)
					continue
				}
				watcher.watchedFiles.Store(file, fs)
				if triggerChange {
					watcher.changeDetected(file)
				}
			}
		}
	}
}

func (watcher *pollWatcher) checkGlobs(delay time.Duration) {
	for {
		watcher.addAllFromGlobs(true)

		select {
		case <-watcher.done:
			return
		case <-time.After(delay * time.Millisecond):
		}
	}
}

func (watcher *pollWatcher) checkFiles(delay time.Duration) {
	for {
		for file, fs := range watcher.watchedFiles.Range() {
			newFs, changed, err := util.HasChanged(file, fs)
			if err != nil {
				log.Println("Error:", err)
			}

			if changed {
				if newFs == nil {
					watcher.watchedFiles.Delete(file)
				} else {
					watcher.watchedFiles.Store(file, newFs)
					watcher.changeDetected(file)
				}
			}
		}

		select {
		case <-watcher.done:
			return
		case <-time.After(delay * time.Millisecond):
		}
	}
}

func (watcher *pollWatcher) changeDetected(filePath string) {
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
