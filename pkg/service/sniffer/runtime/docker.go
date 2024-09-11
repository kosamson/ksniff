package runtime

import (
	"fmt"

	"ksniff/utils"
)

type DockerBridge struct {
	TcpdumpContainerName string
	SocketPath           string
}

func NewDockerBridge() *DockerBridge {
	return &DockerBridge{}
}

func (d *DockerBridge) NeedsPid() bool {
	return false
}

func (d *DockerBridge) BuildInspectCommand(string) []string {
	panic("Docker doesn't need this implemented")
}

func (d *DockerBridge) ExtractPid(inspection string) (*string, error) {
	panic("Docker doesn't need this implemented")
}

func (d *DockerBridge) BuildTcpdumpCommand(args TcpDumpArguments) []string {
	if d.TcpdumpContainerName == "" {
		d.TcpdumpContainerName = "ksniff-container-" + utils.GenerateRandomString(8)
	}

	d.SocketPath = args.SocketPath

	containerNameFlag := fmt.Sprintf("--name=%s", d.TcpdumpContainerName)

	command := []string{
		"docker",
		"--host",
		"unix://" + args.SocketPath,
		"run",
		"--rm",
		"--log-driver",
		"none",
		containerNameFlag,
		fmt.Sprintf("--net=container:%s", *args.ContainerId),
		args.TcpdumpImage,
		"-i",
		args.NetInterface,
		"-U",
		"-w",
		"-",
		args.Filter,
	}

	return command
}

func (d *DockerBridge) BuildCleanupCommand() []string {
	return []string{
		"docker",
		"--host",
		"unix://" + d.SocketPath,
		"rm", "-f", d.TcpdumpContainerName,
	}
}

func (d *DockerBridge) GetDefaultImage() string {
	return "docker"
}

func (d *DockerBridge) GetDefaultTCPImage() string {
	return "maintained/tcpdump"
}

func (d *DockerBridge) GetDefaultSocketPath() string {
	return "/var/run/docker.sock"
}
