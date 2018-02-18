package wado

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/mattn/go-shellwords"
	"github.com/mktange/wado/internal/pkg/util"
)

// CmdRunner can maintain and run a chain of commands
type CmdRunner interface {
	Restart() error
	Start() error
	Kill() error
	SetWriter(writer io.Writer)
	Wait() error
	GetProcess() *os.Process
	GetCommand() []string
}

type cmdDef struct {
	bin  string
	args []string
}

type cmdRun struct {
	bin    string
	args   []string
	cmd    *exec.Cmd
	writer io.Writer
	done   chan error
	mutex  *sync.Mutex
}

// NewCmdRunner creates a new runner for a full command string
func NewCmdRunner(cmd string) (CmdRunner, error) {
	words, err := shellwords.Parse(cmd)
	if err != nil {
		return nil, err
	}
	return NewCmdRunnerBinArgs(words[0], words[1:]...), nil
}

// NewCmdRunnerBinArgs creates a new runner for a command from a binary path and arguments
func NewCmdRunnerBinArgs(bin string, args ...string) CmdRunner {
	return &cmdRun{
		bin:    bin,
		args:   args,
		writer: ioutil.Discard,
		mutex:  &sync.Mutex{},
	}
}

func (r *cmdRun) GetCommand() []string {
	return append([]string{r.bin}, r.args...)
}

func (r *cmdRun) GetProcess() *os.Process {
	if r.isCmdSet() {
		return r.cmd.Process
	}
	return nil
}

func (r *cmdRun) isCmdSet() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.cmd != nil
}

func (r *cmdRun) setCmd(cmd *exec.Cmd) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.cmd = cmd
}

func (r *cmdRun) makeDone() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.done = make(chan error)
}

func (r *cmdRun) closeDone() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	close(r.done)
}

func (r *cmdRun) Restart() error {
	if r.isCmdSet() {
		err := r.Kill()
		if err != nil {
			return err
		}
	}
	return r.Start()
}

func (r *cmdRun) SetWriter(writer io.Writer) {
	r.writer = writer
}

func (r *cmdRun) Wait() error {
	if !r.isCmdSet() {
		return nil
	}
	err := <-r.done
	return err
}

func (r *cmdRun) Start() error {
	if r.isCmdSet() {
		return errors.New("already running")
	}

	r.setCmd(exec.Command(r.bin, r.args...))
	r.makeDone()
	util.SetupCmd(r.cmd)

	// Pipe stdout and stderr to writer
	stdout, err := r.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := r.cmd.StderrPipe()
	if err != nil {
		return err
	}

	go io.Copy(r.writer, stdout)
	go io.Copy(r.writer, stderr)

	// Start cmd
	err = r.cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		r.cmd.Wait()
		r.setCmd(nil)
		r.closeDone()
	}()
	return nil
}

func (r *cmdRun) Kill() error {
	if r.GetProcess() != nil {

		if runtime.GOOS == "windows" {
			// Kill immediatly on windows
			err := util.HardKill(r.cmd.Process.Pid)
			if err != nil {
				return err
			}

		} else {

			// Try to stop it with an interrupt on unix
			if err := r.cmd.Process.Signal(os.Interrupt); err != nil {
				return err
			}

			// Wait for the process to finish, or kill after 3 seconds
			select {
			case <-time.After(2 * time.Second):
				if err := r.cmd.Process.Kill(); err != nil {
					log.Println("Failed to kill process:", err)
				}
			case <-r.done:
			}
		}

		// Wait for process to actually stop
		for {
			if r.GetProcess() == nil {
				break
			}
			if !util.IsProcessRunning(r.cmd.Process.Pid) {
				break
			}
			<-time.After(10 * time.Millisecond)
		}
	}

	return nil
}
