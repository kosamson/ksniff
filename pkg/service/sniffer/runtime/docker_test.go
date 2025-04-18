package runtime_test

import (
	"testing"

	"ksniff/pkg/service/sniffer/runtime"

	"github.com/stretchr/testify/assert"
)

func TestNeedsPid_Docker(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewDockerBridge()
	assert.Equal(t, false, bridge.NeedsPid())
}

func TestGetDefaultImage_Docker(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewDockerBridge()
	assert.Equal(t, "docker", bridge.GetDefaultImage())
}

func TestGetDefaultTcpImage_Docker(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewDockerBridge()
	assert.Equal(t, "maintained/tcpdump", bridge.GetDefaultTCPImage())
}

func TestGetDefaultSocketPath_Docker(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewDockerBridge()
	assert.Equal(t, "/var/run/docker.sock", bridge.GetDefaultSocketPath())
}

func TestExtractPid_Docker(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewDockerBridge()
	assert.Panics(t, func() { bridge.ExtractPid("") })
}

func TestBuildInspectCommand_Docker(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewDockerBridge()
	assert.Panics(t, func() { bridge.BuildInspectCommand("") })
}

func TestBuildTcpDumpCommand_Docker(t *testing.T) {
	t.Parallel()

	args := runtime.TcpDumpArguments{
		ContainerId:  stringPtr("container"),
		NetInterface: "eth0",
		Filter:       "filter",
		Pid:          stringPtr("pid"),
		SocketPath:   "/path",
		TcpdumpImage: "docker.io/kosamson/tcpdump",
	}

	expectedOutput := []string{
		"docker",
		"--host",
		"unix:///path",
		"run",
		"--rm",
		"--log-driver",
		"none",
		"--name=ksniff-container-abcd",
		"--net=container:container",
		"docker.io/kosamson/tcpdump",
		"-i",
		"eth0",
		"-U",
		"-w",
		"-",
		"filter",
	}

	bridge := runtime.NewDockerBridge()

	bridge.TcpdumpContainerName = "ksniff-container-abcd"

	tcpDumpCommand := bridge.BuildTcpdumpCommand(args)

	assert.Equal(t, expectedOutput, tcpDumpCommand)
	assert.Equal(t, args.SocketPath, bridge.SocketPath)
}

func TestBuildCleanupCommand_Docker(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewDockerBridge()

	bridge.TcpdumpContainerName = "container"
	bridge.SocketPath = "/path"

	expectedCommand := []string{
		"docker",
		"--host",
		"unix://" + bridge.SocketPath,
		"rm", "-f", bridge.TcpdumpContainerName,
	}

	bridge.BuildCleanupCommand()
	assert.Equal(
		t,
		expectedCommand,
		bridge.BuildCleanupCommand(),
		"container cleanup command doesn't match",
	)
}
