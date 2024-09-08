package runtime_test

import (
	"testing"

	"ksniff/pkg/service/sniffer/runtime"

	"github.com/stretchr/testify/assert"
)

func TestNeedsPid_Containerd(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewContainerdBridge()
	assert.Equal(t, false, bridge.NeedsPid())
}

func TestGetDefaultImage_Containerd(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewContainerdBridge()
	assert.Equal(t, "docker.io/hamravesh/ksniff-helper:v3", bridge.GetDefaultImage())
}

func TestGetDefaultTcpImage_Containerd(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewContainerdBridge()
	assert.Equal(t, "docker.io/maintained/tcpdump:latest", bridge.GetDefaultTCPImage())
}

func TestGetDefaultSocketPath_Containerd(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewContainerdBridge()
	assert.Equal(t, "/run/containerd/containerd.sock", bridge.GetDefaultSocketPath())
}

func TestBuildInspectCommand_Containerd(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewContainerdBridge()

	assert.Panics(t, func() {
		bridge.BuildInspectCommand("test-container")
	})
}

func TestExtractPid_Containerd(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewContainerdBridge()

	assert.Panics(t, func() {
		bridge.ExtractPid("")
	})
}

func TestBuildTcpdumpCommand_Containerd(t *testing.T) {
	t.Parallel()

	args := runtime.TcpDumpArguments{
		SocketPath:   "/run/custom/containerd.sock",
		ContainerId:  stringPtr("my-container"),
		NetInterface: "iface",
		Filter:       "tcp",
		TcpdumpImage: "docker.io/kosamson/tcpdump",
	}

	shellScript := `
    set -ex
    export CONTAINERD_SOCKET="/run/custom/containerd.sock"
    export CONTAINERD_NAMESPACE="k8s.io"
    export CONTAINER_RUNTIME_ENDPOINT="unix:///host${CONTAINERD_SOCKET}"
    export IMAGE_SERVICE_ENDPOINT=${CONTAINER_RUNTIME_ENDPOINT}
    crictl pull docker.io/kosamson/tcpdump >/dev/null
    netns=$(crictl inspect my-container | jq '.info.runtimeSpec.linux.namespaces[] | select(.type == "network") | .path' | tr -d '"')
    exec chroot /host ctr -a ${CONTAINERD_SOCKET} run --rm --with-ns "network:${netns}" docker.io/kosamson/tcpdump ksniff-container-abcd tcpdump -i iface -U -w - tcp`

	expectedOutput := []string{
		"/bin/sh",
		"-c",
		shellScript,
	}

	bridge := runtime.NewContainerdBridge()

	bridge.TcpdumpContainerName = "ksniff-container-abcd"

	tcpDumpCommand := bridge.BuildTcpdumpCommand(args)

	assert.Equal(t, expectedOutput, tcpDumpCommand)
	assert.Equal(t, args.SocketPath, bridge.SocketPath)
}

func TestBuildCleanupCommand_Containerd(t *testing.T) {
	t.Parallel()

	shellScript := `
    set -ex
    export CONTAINERD_SOCKET="/run/custom/containerd.sock"
    export CONTAINERD_NAMESPACE="k8s.io"
    export CONTAINER_ID="ksniff-container-abcd"
    chroot /host ctr -a ${CONTAINERD_SOCKET} task kill -s SIGKILL ${CONTAINER_ID}`

	expectedOutput := []string{
		"/bin/sh",
		"-c",
		shellScript,
	}

	bridge := runtime.NewContainerdBridge()

	bridge.TcpdumpContainerName = "ksniff-container-abcd"
	bridge.SocketPath = "/run/custom/containerd.sock"

	cleanupCommand := bridge.BuildCleanupCommand()

	assert.Equal(t, expectedOutput, cleanupCommand)
}
