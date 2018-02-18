package wado

import (
	"strings"
	"testing"
	"time"

	"github.com/mktange/wado/internal/pkg/syncimpls"
	"github.com/mktange/wado/internal/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getCounterRunner(counterArgs ...string) (CmdRunner, error) {
	args := append(
		[]string{util.GetCounterRunCmd()},
		counterArgs...,
	)
	return NewCmdRunner(strings.Join(args, " "))
}

func Test_Start(t *testing.T) {
	runner, err := getCounterRunner("2")
	require.NoError(t, err)

	buffer := syncimpls.NewSyncBuffer()
	runner.SetWriter(buffer)

	err = runner.Start()
	require.NoError(t, err)

	err = runner.Wait()
	require.NoError(t, err)

	assert.Equal(t, "Starting count\nCounter: 0\nCounter: 1\n", buffer.String())
}

func Test_Kill(t *testing.T) {
	runner, err := getCounterRunner()
	require.NoError(t, err)

	buffer := syncimpls.NewSyncBuffer()
	runner.SetWriter(buffer)

	err = runner.Start()
	require.NoError(t, err)
	assert.NotNil(t, runner.GetProcess())
	pid := runner.GetProcess().Pid

	util.WaitForChange(buffer)

	err = runner.Kill()
	require.NoError(t, err)

	assert.Nil(t, runner.GetProcess())
	assert.False(t, util.IsProcessRunning(pid))

	before := buffer.String()
	assert.True(t, len(before) > 0, "Buffer size should be greater than 0")

	<-time.After(30 * time.Millisecond)

	assert.Equal(t, before, buffer.String())
}

func Test_Restart(t *testing.T) {
	runner, err := getCounterRunner()
	require.NoError(t, err)

	buffer := syncimpls.NewSyncBuffer()
	runner.SetWriter(buffer)

	err = runner.Start()
	require.NoError(t, err)

	util.WaitForChange(buffer)

	err = runner.Restart()
	require.NoError(t, err)

	util.WaitForChange(buffer)

	assert.Equal(t, 2, strings.Count(buffer.String(), "Starting count"),
		"Expected program to have been started twice")
}

func Test_ManyRestarts(t *testing.T) {
	runner, err := getCounterRunner("2")
	require.NoError(t, err)

	buffer := syncimpls.NewSyncBuffer()
	runner.SetWriter(buffer)

	for i := 0; i < 50; i++ {
		err = runner.Start()
		require.NoError(t, err)
		err = runner.Wait()
		require.NoError(t, err)
	}

}
