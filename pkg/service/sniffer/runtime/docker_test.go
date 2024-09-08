package runtime_test

import (
	"testing"

	"ksniff/pkg/service/sniffer/runtime"

	"github.com/stretchr/testify/assert"
)

func TestExtractPid(t *testing.T) {
	bridge := runtime.NewDockerBridge()
	assert.Panics(t, func() { bridge.ExtractPid("") })
}

func TestInspectCommand(t *testing.T) {
	bridge := runtime.NewDockerBridge()
	assert.Panics(t, func() { bridge.BuildInspectCommand("") })
}

func TestPrivilegedPodName(t *testing.T) {
	bridge := runtime.NewDockerBridge()
	args := runtime.TcpDumpArguments{
		ContainerId:  stringPtr("container"),
		NetInterface: "eth0",
		Filter:       "filter",
		Pid:          stringPtr("pid"),
		SocketPath:   "/path",
		TcpdumpImage: bridge.GetDefaultTCPImage(),
	}
	bridge.BuildTcpdumpCommand(args)
	assert.NotEqual(t, "", bridge.TcpdumpContainerName, "tcpdumpContainerName should have been set")
}

func TestCleanupCommand(t *testing.T) {
	bridge := runtime.NewDockerBridge()

	args := runtime.TcpDumpArguments{
		ContainerId:  stringPtr("container"),
		NetInterface: "eth0",
		Filter:       "filter",
		Pid:          stringPtr("pid"),
		SocketPath:   "/path",
		TcpdumpImage: bridge.GetDefaultTCPImage(),
	}

	bridge.BuildTcpdumpCommand(args)
	assert.Equal(t,
		[]string{"docker", "--host", "unix://" + args.SocketPath, "rm", "-f", bridge.TcpdumpContainerName},
		bridge.BuildCleanupCommand(),
		"container cleanup command doesn't match")
}
