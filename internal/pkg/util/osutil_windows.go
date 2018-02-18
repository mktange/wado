package util

import (
	"os"
	"os/exec"
	"strconv"
)

// HardKill kills the process witht he given PID for real
func HardKill(pid int) error {
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(pid))
	return kill.Run()
}

// IsProcessRunning checks if a given process is running
func IsProcessRunning(pid int) bool {
	_, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return true
}

// SetupCmd sets up a runnable/killable command in the OS
func SetupCmd(cmd *exec.Cmd) {
}
