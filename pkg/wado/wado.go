package wado

import (
	"log"
	"os"
	"sync"
	"time"
)

// Instance listens for changes via the watcher and issues commands to the runner
type Instance interface {
	Kill()
}

// Config holds information regarding a specific watcher configuration
type Config struct {
	Name         string   `yaml:"name,omitempty"`
	IncludeGlobs []string `yaml:"include,omitempty"`
	ExcludeGlobs []string `yaml:"exclude,omitempty"`
	Cmds         []string `yaml:"cmds,omitempty"`
	MinDelay     int      `yaml:"minDelay,omitempty"`
}

type wadoInstance struct {
	name      string
	watcher   Watcher
	cmdChain  CmdChain
	minDelay  time.Duration
	lastStart time.Time
	mutex     *sync.Mutex
}

// New creates a new wado instance for the given configuration
func New(config Config) (Instance, error) {
	name := config.Name
	if name == "" {
		name = "Wado"
	}

	watcher, err := NewPollWatcher(config.IncludeGlobs, config.ExcludeGlobs)
	if err != nil {
		return nil, err
	}

	cmdChain, err := NewCmdChain(config.Cmds...)
	if err != nil {
		return nil, err
	}
	cmdChain.SetWriter(os.Stdout)

	minDelay := config.MinDelay
	if minDelay <= 0 {
		minDelay = 50
	}

	wado := &wadoInstance{
		name:      name,
		watcher:   watcher,
		cmdChain:  cmdChain,
		minDelay:  time.Duration(minDelay) * time.Millisecond,
		lastStart: time.Unix(0, 0),
		mutex:     &sync.Mutex{},
	}

	watcher.AddCallback(wado.changeEvent)
	err = wado.start()
	if err != nil {
		return nil, err
	}

	// fmt.Printf("[%v] Currently watching %v files.\n", wado.name, watcher.FileCount())

	return wado, nil
}

func (m *wadoInstance) start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.lastStart = time.Now()
	return m.cmdChain.Start()
}

func (m *wadoInstance) changeEvent(filePath string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if time.Since(m.lastStart) >= m.minDelay {
		err := m.cmdChain.Restart()
		if err != nil {
			log.Println("Error while restarting command chain:", err)
			return
		}
		m.lastStart = time.Now()
	}
}

// Kill stops the current running command chain and closes the watcher
func (m *wadoInstance) Kill() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		m.cmdChain.Kill()
		wg.Done()
	}()
	go func() {
		m.watcher.Close()
		wg.Done()
	}()
	wg.Wait()
}
