// +build !windows

package util

import (
	"os"
	"os/exec"
	"syscall"
)

// HardKill kills the process witht he given PID for real
func HardKill(pid int) error {
	return syscall.Kill(-pid, syscall.SIGKILL)
}

// IsProcessRunning checks if a given process is running
func IsProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

// SetupCmd sets up a runnable/killable command in the OS
func SetupCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
