package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

type hasLen interface {
	Len() int
}

// WaitForChange blocks until the length of the input has changed
func WaitForChange(buffer hasLen) {
	start := buffer.Len()
	for {
		<-time.After(10 * time.Millisecond)
		if buffer.Len() != start {
			break
		}
	}
}

// WaitForStabilize blocks until the length of the input is stable for 10 ms
func WaitForStabilize(buffer hasLen) {
	previous := buffer.Len()
	for {
		<-time.After(10 * time.Millisecond)
		if buffer.Len() == previous {
			break
		} else {
			previous = buffer.Len()
		}
	}
}

// WaitForMessage returns true if a message appears on the given channel within 200ms
func WaitForMessage(t *testing.T, ch <-chan string) bool {
	select {
	case <-ch:
		return true
	case <-time.After(200 * time.Millisecond):
		return false
	}
}

func getTestFixtureCmd(fixtureName string) string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	pathToFixture := filepath.Join(basepath, "../../../test/fixtures", fixtureName)

	binPath := filepath.Join(pathToFixture, "bin", fixtureName)
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	err := os.MkdirAll(filepath.Dir(binPath), os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("could not create test fixture bin folder: %v", err))
	}

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		cmd := exec.Command("go", "build", "-o", binPath, filepath.Join(pathToFixture, "main.go"))
		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		err = cmd.Wait()
		if err != nil {
			panic(err)
		}
	}

	return fmt.Sprintf("'%v'", binPath)
}

// GetCounterRunCmd returns the path to run the counter test fixture
func GetCounterRunCmd() string {
	return getTestFixtureCmd("counter")
}
