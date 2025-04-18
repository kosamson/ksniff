package runtime

import "fmt"

var SupportedContainerRuntimes = []string{
	"docker",
	"cri-o",
	"containerd",
}

type TcpDumpArguments struct {
	ContainerId  *string
	NetInterface string
	Filter       string
	Pid          *string
	SocketPath   string
	TcpdumpImage string
}

type ContainerRuntimeBridge interface {
	NeedsPid() bool
	BuildInspectCommand(containerId string) []string
	ExtractPid(inspection string) (*string, error)
	BuildTcpdumpCommand(TcpDumpArguments) []string
	BuildCleanupCommand() []string
	GetDefaultImage() string
	GetDefaultTCPImage() string
	GetDefaultSocketPath() string
}

func NewContainerRuntimeBridge(runtimeName string) ContainerRuntimeBridge {
	switch runtimeName {
	case "docker":
		return NewDockerBridge()
	case "cri-o":
		return NewCrioBridge()
	case "containerd":
		return NewContainerdBridge()
	default:
		panic(fmt.Sprintf("Unable to build bridge to %s", runtimeName))
	}
}
