package sniffer_test

import (
	"testing"

	"ksniff/kube"
	"ksniff/pkg/config"
	"ksniff/pkg/service/sniffer"
	"ksniff/pkg/service/sniffer/runtime"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

type (
	NopRuntimeBridge     struct{}
	ModularRuntimeBridge struct {
		needsPid             func() bool
		buildInspectCommand  func(string) []string
		extractPid           func(string) (*string, error)
		buildTcpdumpCommand  func(runtime.TcpDumpArguments) []string
		buildCleanupCommand  func() []string
		getDefaultImage      func() string
		getDefaultTCPImage   func() string
		getDefaultSocketPath func() string
	}
)

var (
	_ runtime.ContainerRuntimeBridge = NopRuntimeBridge{}
	_ runtime.ContainerRuntimeBridge = ModularRuntimeBridge{}
)

func (bridge NopRuntimeBridge) NeedsPid() bool {
	return true
}

func (bridge NopRuntimeBridge) BuildInspectCommand(containerId string) []string {
	return []string{
		"inspect",
	}
}

func (bridge NopRuntimeBridge) ExtractPid(inspection string) (*string, error) {
	return stringPtr("1234"), nil
}

func (bridge NopRuntimeBridge) BuildTcpdumpCommand(args runtime.TcpDumpArguments) []string {
	return []string{
		"tcpdump",
	}
}

func (bridge NopRuntimeBridge) BuildCleanupCommand() []string {
	return []string{}
}

func (bridge NopRuntimeBridge) GetDefaultImage() string {
	return "custom-image"
}

func (bridge NopRuntimeBridge) GetDefaultTCPImage() string {
	return "tcpdump-image"
}

func (bridge NopRuntimeBridge) GetDefaultSocketPath() string {
	return "/custom.sock"
}

func (bridge ModularRuntimeBridge) NeedsPid() bool {
	return true
}

func (bridge ModularRuntimeBridge) BuildInspectCommand(containerId string) []string {
	return bridge.buildInspectCommand(containerId)
}

func (bridge ModularRuntimeBridge) ExtractPid(inspection string) (*string, error) {
	return bridge.extractPid(inspection)
}

func (bridge ModularRuntimeBridge) BuildTcpdumpCommand(args runtime.TcpDumpArguments) []string {
	return bridge.buildTcpdumpCommand(args)
}

func (bridge ModularRuntimeBridge) BuildCleanupCommand() []string {
	return bridge.buildCleanupCommand()
}

func (bridge ModularRuntimeBridge) GetDefaultImage() string {
	return bridge.getDefaultImage()
}

func (bridge ModularRuntimeBridge) GetDefaultTCPImage() string {
	return bridge.getDefaultTCPImage()
}

func (bridge ModularRuntimeBridge) GetDefaultSocketPath() string {
	return bridge.getDefaultSocketPath()
}

func TestSetup_Privileged(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	testCases := []struct {
		name       string
		settings   *config.KsniffSettings
		apiservice kube.KubernetesApiService
		bridge     runtime.ContainerRuntimeBridge
		expectErr  bool
	}{
		{
			name:       "happy path",
			settings:   &config.KsniffSettings{},
			apiservice: NopKubernetesApiService{},
			bridge:     NopRuntimeBridge{},
			expectErr:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			svc := sniffer.NewPrivilegedPodRemoteSniffingService(
				testCase.settings,
				testCase.apiservice,
				testCase.bridge,
			)

			err := svc.Setup()

			if testCase.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestStart_Privileged(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	testCases := []struct {
		name       string
		settings   *config.KsniffSettings
		apiservice kube.KubernetesApiService
		bridge     runtime.ContainerRuntimeBridge
		expectErr  bool
	}{
		{
			name:       "happy path",
			settings:   &config.KsniffSettings{},
			apiservice: NopKubernetesApiService{},
			bridge:     NopRuntimeBridge{},
			expectErr:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			svc := sniffer.NewPrivilegedPodRemoteSniffingService(
				testCase.settings,
				testCase.apiservice,
				testCase.bridge,
			)

			svc.PrivilegedPod = &v1.Pod{}

			err := svc.Start(nil)

			if testCase.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestCleanup_Privileged(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	testCases := []struct {
		name       string
		settings   *config.KsniffSettings
		apiservice kube.KubernetesApiService
		bridge     runtime.ContainerRuntimeBridge
		expectErr  bool
	}{
		{
			name:       "happy path",
			settings:   &config.KsniffSettings{},
			apiservice: NopKubernetesApiService{},
			bridge:     NopRuntimeBridge{},
			expectErr:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			svc := sniffer.NewPrivilegedPodRemoteSniffingService(
				testCase.settings,
				testCase.apiservice,
				testCase.bridge,
			)

			svc.PrivilegedPod = &v1.Pod{}

			err := svc.Cleanup()

			if testCase.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
