package wado

import (
	"errors"
	"io"
	"log"
	"sync"
)

// CmdChain maintains a chain of commands and runs them in sequence
type CmdChain interface {
	SetWriter(io.Writer)
	Start() error
	Restart() error
	Kill()
	Wait()
	IsRunning() bool
}

type cmdChain struct {
	runners    []CmdRunner
	isRunning  bool
	shouldKill chan bool
	isDone     chan error
	mutex      *sync.Mutex
}

// NewCmdChain creates a new CmdChain based on the given list of commands
func NewCmdChain(cmds ...string) (CmdChain, error) {
	runners := []CmdRunner{}
	for _, cmd := range cmds {
		cmdRunner, err := NewCmdRunner(cmd)
		if err != nil {
			return nil, err
		}
		runners = append(runners, cmdRunner)
	}

	return &cmdChain{
		runners:   runners,
		isRunning: false,
		mutex:     &sync.Mutex{},
	}, nil
}

// SetWriter sets the writer for all commands in the chain
func (c *cmdChain) SetWriter(writer io.Writer) {
	for _, runner := range c.runners {
		runner.SetWriter(writer)
	}
}

// IsRunning returns true if the chain is currently running
func (c *cmdChain) IsRunning() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isRunning
}

func (c *cmdChain) SetIsRunning(running bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.isRunning = running
}

// Start starts the command chain
func (c *cmdChain) Start() error {
	if c.IsRunning() {
		return errors.New("already running")
	}

	c.isDone = make(chan error)
	c.SetIsRunning(true)
	c.shouldKill = make(chan bool, 50)
	go c.startChain()
	return nil
}

// Restart restarts the command chain
func (c *cmdChain) Restart() error {
	var err error
	if c.IsRunning() {
		c.Kill()
		if err != nil {
			return err
		}
	}
	return c.Start()
}

// Kill kills the currently running command and stops the execution of the following ones
func (c *cmdChain) Kill() {
	if c.IsRunning() {
		if c.shouldKill != nil && len(c.shouldKill) == 0 {
			c.shouldKill <- true
		}
		<-c.isDone
	}
}

// Wait waits until all the commands in the chain have finished executing
func (c *cmdChain) Wait() {
	if c.IsRunning() {
		<-c.isDone
	}
}

// Main method for executing the chain of commands
func (c *cmdChain) startChain() {
	var err error
	wasKilled := false
	for _, runner := range c.runners {
		if wasKilled {
			break
		}

		err = runner.Start()
		if err != nil {
			log.Printf("Error: could not run the command (%v): %v\n", runner.GetCommand(), err)
		}

		done := make(chan error)
		go func() {
			done <- runner.Wait()
		}()

		select {
		case <-c.shouldKill:
			err = runner.Kill()
			if err != nil {
				log.Printf("Error: waiting on command failed (%v): %v\n", runner.GetCommand(), err)
			}
			wasKilled = true
			break
		case err := <-done:
			if err != nil {
				log.Printf("Error: waiting on command failed (%v): %v\n", runner.GetCommand(), err)
			}
		}
	}

	c.SetIsRunning(false)
	close(c.shouldKill)
	close(c.isDone)
}
