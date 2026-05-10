package listenaddr

import (
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAvailableListenAddr_succeeds(t *testing.T) {
	t.Setenv("PORT", "")
	addr, err := AvailableListenAddr()
	require.NoError(t, err)
	require.NotEmpty(t, addr)
	assert.Regexp(t, `^:\d+$`, addr)

	ln, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
}

func TestAvailableListenAddr_respectsPORT(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	_, portStr, splitErr := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, splitErr)
	_ = ln.Close()

	t.Setenv("PORT", portStr)
	addr, err := AvailableListenAddr()
	require.NoError(t, err)
	assert.Equal(t, ":"+portStr, addr)
}

func TestAvailableListenAddr_invalidPORTFallsBack(t *testing.T) {
	t.Setenv("PORT", "not-a-port")
	addr, err := AvailableListenAddr()
	require.NoError(t, err)
	p, atoiErr := strconv.Atoi(addr[1:])
	require.NoError(t, atoiErr)
	assert.GreaterOrEqual(t, p, 8000)
	assert.LessOrEqual(t, p, 8999)
}

func TestPortFromEnv(t *testing.T) {
	t.Setenv("PORT", "")
	assert.Equal(t, 0, portFromEnv())

	t.Setenv("PORT", " 9001 ")
	assert.Equal(t, 9001, portFromEnv())

	t.Setenv("PORT", "0")
	assert.Equal(t, 0, portFromEnv())

	t.Setenv("PORT", "65536")
	assert.Equal(t, 0, portFromEnv())
}
