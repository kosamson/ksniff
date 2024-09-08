package runtime

import (
	"fmt"

	"ksniff/utils"
)

type ContainerdBridge struct {
	TcpdumpContainerName string
	SocketPath           string
}

func NewContainerdBridge() *ContainerdBridge {
	return &ContainerdBridge{}
}

func (d ContainerdBridge) NeedsPid() bool {
	return false
}

func (d ContainerdBridge) BuildInspectCommand(string) []string {
	panic("Containerd doesn't need this implemented")
}

func (d ContainerdBridge) ExtractPid(inspection string) (*string, error) {
	panic("Containerd doesn't need this implemented")
}

func (d ContainerdBridge) GetDefaultSocketPath() string {
	return "/run/containerd/containerd.sock"
}

func (d *ContainerdBridge) BuildTcpdumpCommand(args TcpDumpArguments) []string {
	if d.TcpdumpContainerName == "" {
		d.TcpdumpContainerName = "ksniff-container-" + utils.GenerateRandomString(nil, 8)
	}

	d.SocketPath = args.SocketPath
	tcpdumpCommand := fmt.Sprintf("tcpdump -i %s -U -w - %s", args.NetInterface, args.Filter)
	shellScript := fmt.Sprintf(`
    set -ex
    export CONTAINERD_SOCKET="%s"
    export CONTAINERD_NAMESPACE="k8s.io"
    export CONTAINER_RUNTIME_ENDPOINT="unix:///host${CONTAINERD_SOCKET}"
    export IMAGE_SERVICE_ENDPOINT=${CONTAINER_RUNTIME_ENDPOINT}
    crictl pull %s >/dev/null
    netns=$(crictl inspect %s | jq '.info.runtimeSpec.linux.namespaces[] | select(.type == "network") | .path' | tr -d '"')
    exec chroot /host ctr -a ${CONTAINERD_SOCKET} run --rm --with-ns "network:${netns}" %s %s %s`, d.SocketPath, args.TcpdumpImage, *args.ContainerId, args.TcpdumpImage, d.TcpdumpContainerName, tcpdumpCommand)
	command := []string{"/bin/sh", "-c", shellScript}
	return command
}

func (d *ContainerdBridge) BuildCleanupCommand() []string {
	shellScript := fmt.Sprintf(`
    set -ex
    export CONTAINERD_SOCKET="%s"
    export CONTAINERD_NAMESPACE="k8s.io"
    export CONTAINER_ID="%s"
    chroot /host ctr -a ${CONTAINERD_SOCKET} task kill -s SIGKILL ${CONTAINER_ID}`, d.SocketPath, d.TcpdumpContainerName)
	command := []string{"/bin/sh", "-c", shellScript}
	return command
}

func (d ContainerdBridge) GetDefaultImage() string {
	return "docker.io/hamravesh/ksniff-helper:v3"
}

func (d *ContainerdBridge) GetDefaultTCPImage() string {
	return "docker.io/maintained/tcpdump:latest"
}
