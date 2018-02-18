package wado

import (
	"testing"

	"github.com/mktange/wado/internal/pkg/syncimpls"
	"github.com/mktange/wado/internal/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CmdChain(t *testing.T) {
	chain, err := NewCmdChain("echo Foo", "echo Bar")
	require.NoError(t, err)

	buffer := syncimpls.NewSyncBuffer()
	chain.SetWriter(buffer)

	chain.Start()
	chain.Wait()

	assert.Equal(t, "Foo\nBar\n", buffer.String())
}

func Test_CmdChainKill(t *testing.T) {
	chain, err := NewCmdChain(
		"echo Before",
		util.GetCounterRunCmd(),
		"echo After",
	)
	require.NoError(t, err)

	buffer := syncimpls.NewSyncBuffer()
	chain.SetWriter(buffer)

	chain.Start()

	util.WaitForChange(buffer)
	util.WaitForChange(buffer)

	chain.Kill()

	assert.Contains(t, buffer.String(), "Before")
	assert.Contains(t, buffer.String(), "Counter")
	assert.NotContains(t, buffer.String(), "After")
}
