package runtime_test

import (
	"testing"

	"ksniff/pkg/service/sniffer/runtime"

	"github.com/stretchr/testify/assert"
)

const (
	CRICTL_INSPECT_NO_PID_117 = `
{
  "status": {},
  "sandboxId": "0bba370d1a514cd5242f11b707dabc8cc54d0a653e39a9c787ec5be7e80ec887"
}
`

	CRICTL_INSPECT_WITH_PID_117 = `
{
  "status": {},
  "pid": 69417,
  "sandboxId": "0bba370d1a514cd5242f11b707dabc8cc54d0a653e39a9c787ec5be7e80ec887"
}
`

	CRICTL_INSPECT_WITH_PID_118 = `
{
  "status": {},
  "info": {
    "sandboxID": "549eb241ba685900fc152501be3c2c31b19e9d649c01f00496b58375e570da52",
    "pid": 827137
    }
  }
`

	CRICTL_INSPECT_NO_PID_118 = `
{
  "status": {},
  "info": {
    "sandboxID": "549eb241ba685900fc152501be3c2c31b19e9d649c01f00496b58375e570da52",
	}
}
`
)

func TestNeedsPid_Crio(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewCrioBridge()
	assert.Equal(t, true, bridge.NeedsPid())
}

func TestBuildCleanupCommand_Crio(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewCrioBridge()
	assert.Nil(t, bridge.BuildCleanupCommand())
}

func TestGetDefaultImage_Crio(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewCrioBridge()
	assert.Equal(t, "maintained/tcpdump", bridge.GetDefaultImage())
}

func TestGetDefaultTcpImage_Crio(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewCrioBridge()
	assert.Equal(t, "", bridge.GetDefaultTCPImage())
}

func TestGetDefaultSocketPath_Crio(t *testing.T) {
	t.Parallel()

	bridge := runtime.NewCrioBridge()
	assert.Equal(t, "/var/run/crio/crio.sock", bridge.GetDefaultSocketPath())
}

func TestBuildInspectCommand_Crio(t *testing.T) {
	t.Parallel()

	expectedOutput := []string{
		"chroot",
		"/host",
		"crictl",
		"inspect",
		"--output",
		"json",
		"test-container",
	}

	bridge := runtime.NewCrioBridge()
	inspectCommand := bridge.BuildInspectCommand("test-container")

	assert.Equal(t, expectedOutput, inspectCommand)
}

func TestBuildTcpdumpCommand_Crio(t *testing.T) {
	t.Parallel()

	args := runtime.TcpDumpArguments{
		Pid:          stringPtr("1234"),
		NetInterface: "iface",
		Filter:       "tcp",
	}

	expectedOutput := []string{
		"nsenter",
		"-n",
		"-t",
		"1234",
		"--",
		"tcpdump",
		"-i",
		"iface",
		"-U",
		"-w",
		"-",
		"tcp",
	}

	bridge := runtime.NewCrioBridge()
	tcpDumpCommand := bridge.BuildTcpdumpCommand(args)

	assert.Equal(t, expectedOutput, tcpDumpCommand)
}

func TestExtractPid_Crio(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		inputInspectData string
		expectedPid      *string
		expectErr        bool
	}{
		{
			name:             "empty input",
			inputInspectData: "",
			expectedPid:      nil,
			expectErr:        true,
		},
		{
			name:             "empty json",
			inputInspectData: "{}",
			expectedPid:      nil,
			expectErr:        true,
		},
		{
			name:             "no pid cri-o 1.17",
			inputInspectData: CRICTL_INSPECT_NO_PID_117,
			expectedPid:      nil,
			expectErr:        true,
		},
		{
			name:             "valid pid cri-o 1.17",
			inputInspectData: CRICTL_INSPECT_WITH_PID_117,
			expectedPid:      stringPtr("69417"),
			expectErr:        false,
		},
		{
			name:             "no pid cri-o 1.18",
			inputInspectData: CRICTL_INSPECT_NO_PID_118,
			expectedPid:      nil,
			expectErr:        true,
		},
		{
			name:             "valid pid cri-o 1.18",
			inputInspectData: CRICTL_INSPECT_WITH_PID_118,
			expectedPid:      stringPtr("827137"),
			expectErr:        false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert := assert.New(t)
			bridge := runtime.NewCrioBridge()

			extractedPid, err := bridge.ExtractPid(testCase.inputInspectData)

			if testCase.expectErr {
				assert.NotNil(err)
			} else {
				assert.Nil(err)
			}

			assert.Equal(testCase.expectedPid, extractedPid)
		})
	}
}
