package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type CrioBridge struct{}

func NewCrioBridge() *CrioBridge {
	return &CrioBridge{}
}

func (c *CrioBridge) NeedsPid() bool {
	return true
}

func (c *CrioBridge) BuildInspectCommand(containerId string) []string {
	return []string{
		"chroot",
		"/host",
		"crictl",
		"inspect",
		"--output",
		"json",
		containerId,
	}
}

func (c *CrioBridge) ExtractPid(inspection string) (*string, error) {
	var result map[string]json.RawMessage
	var pid float64
	var err error

	err = json.Unmarshal([]byte(inspection), &result)
	if err != nil {
		return nil, err
	}

	var extractPidFunc func(map[string]json.RawMessage) (float64, error)

	// CRI-O changes the way it reports PID so we have to by dynamic here
	if _, ok := result["pid"]; ok {
		extractPidFunc = extractPidCrio117
	} else if _, ok := result["info"]; ok {
		extractPidFunc = extractPidCrio118
	} else {
		return nil, errors.New("unable to identify CRI-O version")
	}

	pid, err = extractPidFunc(result)
	if err != nil {
		return nil, errors.Wrap(err, "error getting container PID from CRI-O")
	}

	ret := fmt.Sprintf("%.0f", pid)
	return &ret, nil
}

func (c *CrioBridge) BuildTcpdumpCommand(args TcpDumpArguments) []string {
	return []string{
		"nsenter",
		"-n",
		"-t",
		*args.Pid,
		"--",
		"tcpdump",
		"-i",
		args.NetInterface,
		"-U",
		"-w",
		"-",
		args.Filter,
	}
}

func (c *CrioBridge) BuildCleanupCommand() []string {
	return nil // No cleanup needed
}

func (c *CrioBridge) GetDefaultImage() string {
	return "maintained/tcpdump"
}

func (c *CrioBridge) GetDefaultSocketPath() string {
	return "/var/run/crio/crio.sock"
}

// CRI-O 1.17 and older have pid as first-level attribute
func extractPidCrio117(partial map[string]json.RawMessage) (float64, error) {
	resultPid, ok := partial["pid"]
	if !ok {
		return -1, errors.New("pid key not found in cri-o 1.17 inspect result")
	}

	var result float64

	if err := json.Unmarshal(resultPid, &result); err != nil {
		return -1, err
	}

	return result, nil
}

// CRI-O 1.18 and later nest pid under info attribute
func extractPidCrio118(partial map[string]json.RawMessage) (float64, error) {
	partialInfo, ok := partial["info"]
	if !ok {
		return -1, errors.New("info key not found in cri-o 1.18 inspect result")
	}

	var result map[string]interface{}

	err := json.Unmarshal(partialInfo, &result)
	if err != nil {
		return -1, err
	}

	resultPid, ok := result["pid"]
	if !ok {
		return -1, errors.New("pid key not found in cri-o 1.18 inspect result")
	}

	pid, ok := resultPid.(float64)
	if !ok {
		return -1, errors.New("pid value is not of type float64")
	}

	return pid, nil
}

func (d *CrioBridge) GetDefaultTCPImage() string {
	return ""
}
