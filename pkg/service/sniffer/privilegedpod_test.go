package sniffer_test

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"ksniff/kube"
	"ksniff/pkg/config"
	"ksniff/pkg/service/sniffer"
	"ksniff/pkg/service/sniffer/runtime"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
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
		name            string
		settings        *config.KsniffSettings
		apiservice      kube.KubernetesApiService
		bridge          runtime.ContainerRuntimeBridge
		expectErr       bool
		extraAssertions func(*sniffer.PrivilegedPodSnifferService) bool
	}{
		{
			name:       "happy path",
			settings:   &config.KsniffSettings{},
			apiservice: NopKubernetesApiService{},
			bridge:     NopRuntimeBridge{},
			expectErr:  false,
		},
		{
			name: "happy path uses all default values",
			settings: &config.KsniffSettings{
				UseDefaultImage:        true,
				UseDefaultTCPDumpImage: true,
				UseDefaultSocketPath:   true,
			},
			apiservice: NopKubernetesApiService{},
			bridge:     NopRuntimeBridge{},
			expectErr:  false,
			extraAssertions: func(svc *sniffer.PrivilegedPodSnifferService) bool {
				t.Helper()

				nop := NopRuntimeBridge{}

				assert.Equal(nop.GetDefaultImage(), svc.Settings.Image)
				assert.Equal(nop.GetDefaultTCPImage(), svc.Settings.TCPDumpImage)
				assert.Equal(nop.GetDefaultSocketPath(), svc.Settings.SocketPath)

				return true
			},
		},
		{
			name:     "happy path, pid successfully extracted",
			settings: &config.KsniffSettings{},
			apiservice: ModularKubernetesApiService{
				createPrivilegedPod: func(_, _, _, _ string, _ time.Duration, _ string) (*v1.Pod, error) {
					return &corev1.Pod{}, nil
				},
				executeCommand: func(_, _ string, command []string, w io.Writer) (int, error) {
					w.Write([]byte("1234"))

					return 0, nil
				},
			},
			bridge: ModularRuntimeBridge{
				needsPid: func() bool {
					return true
				},
				buildInspectCommand: func(_ string) []string {
					return []string{"inspect"}
				},
				extractPid: func(inspectData string) (*string, error) {
					return stringPtr(inspectData), nil
				},
			},
			expectErr: false,
			extraAssertions: func(svc *sniffer.PrivilegedPodSnifferService) bool {
				t.Helper()

				assert.Equal(*stringPtr("1234"), *svc.TargetProcessId)

				return true
			},
		},
		{
			name:     "sad path, privileged pod failed to create",
			settings: &config.KsniffSettings{},
			apiservice: ModularKubernetesApiService{
				createPrivilegedPod: func(_, _, _, _ string, _ time.Duration, _ string) (*v1.Pod, error) {
					return nil, fmt.Errorf("failed to create privileged pod")
				},
			},
			bridge:    NopRuntimeBridge{},
			expectErr: true,
		},
		{
			name:     "sad path, k8s api svc returns error on ExecuteCommand",
			settings: &config.KsniffSettings{},
			apiservice: ModularKubernetesApiService{
				createPrivilegedPod: func(_, _, _, _ string, _ time.Duration, _ string) (*v1.Pod, error) {
					return &corev1.Pod{}, nil
				},
				executeCommand: func(_, _ string, _ []string, _ io.Writer) (int, error) {
					return 0, fmt.Errorf("error executing command")
				},
			},
			bridge:    NopRuntimeBridge{},
			expectErr: true,
		},
		{
			name:     "sad path, k8s api svc returns non-zero exit code on ExecuteCommand",
			settings: &config.KsniffSettings{},
			apiservice: ModularKubernetesApiService{
				createPrivilegedPod: func(_, _, _, _ string, _ time.Duration, _ string) (*v1.Pod, error) {
					return &corev1.Pod{}, nil
				},
				executeCommand: func(_, _ string, _ []string, _ io.Writer) (int, error) {
					return 1, nil
				},
			},
			bridge:    NopRuntimeBridge{},
			expectErr: true,
		},
		{
			name:       "sad path, failed to extract pid",
			settings:   &config.KsniffSettings{},
			apiservice: NopKubernetesApiService{},
			bridge: ModularRuntimeBridge{
				needsPid: func() bool {
					return true
				},
				buildInspectCommand: func(_ string) []string {
					return []string{}
				},
				extractPid: func(_ string) (*string, error) {
					return nil, fmt.Errorf("failed to extract pid")
				},
			},
			expectErr: true,
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

			if testCase.extraAssertions != nil {
				assert.True(testCase.extraAssertions(svc))
			}

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
		{
			name:     "sad path, k8s api svc ExecuteCommand returns error",
			settings: &config.KsniffSettings{},
			apiservice: ModularKubernetesApiService{
				executeCommand: func(_, _ string, _ []string, _ io.Writer) (int, error) {
					return 0, fmt.Errorf("failed to execute command")
				},
			},
			bridge:    NopRuntimeBridge{},
			expectErr: true,
		},
		{
			name:     "sad path, k8s api svc ExecuteCommand returns non-zero exit code",
			settings: &config.KsniffSettings{},
			apiservice: ModularKubernetesApiService{
				executeCommand: func(_, _ string, _ []string, _ io.Writer) (int, error) {
					return 1, nil
				},
			},
			bridge:    NopRuntimeBridge{},
			expectErr: true,
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
	// unfortunately we cannot parallelize this test
	// because we use t.Setenv for mocking k8sapisvc.DeletePod
	assert := assert.New(t)

	testCases := []struct {
		name            string
		settings        *config.KsniffSettings
		apiservice      kube.KubernetesApiService
		bridge          runtime.ContainerRuntimeBridge
		privilegedPod   *v1.Pod
		expectErr       bool
		extraAssertions func(*sniffer.PrivilegedPodSnifferService) bool
	}{
		{
			name:          "happy path with privileged pod",
			settings:      &config.KsniffSettings{},
			apiservice:    NopKubernetesApiService{},
			bridge:        NopRuntimeBridge{},
			privilegedPod: &v1.Pod{},
			expectErr:     false,
		},
		// we ignore the "happy" sad path(s) where removing priv container succeeds
		// since the operation causes no side effects or different logical paths
		// except for logging
		//
		// (should we be testing logging?)
		{
			name:          "sad path no privileged pod",
			settings:      &config.KsniffSettings{},
			apiservice:    NopKubernetesApiService{},
			bridge:        NopRuntimeBridge{},
			privilegedPod: nil,
			expectErr:     true,
		},
		{
			name:     "'happy' sad path, remove priv container fails, but priv pod still cleaned up",
			settings: &config.KsniffSettings{},
			apiservice: ModularKubernetesApiService{
				executeCommand: func(_, _ string, _ []string, _ io.Writer) (int, error) {
					return -1, fmt.Errorf("failed to execute command")
				},
				deletePod: func(_ string) error {
					t.Setenv("POD_DELETED", "true")

					return nil
				},
			},
			bridge:        NopRuntimeBridge{},
			privilegedPod: &v1.Pod{},
			expectErr:     false,
			extraAssertions: func(svc *sniffer.PrivilegedPodSnifferService) bool {
				t.Helper()

				assert.Equal(os.Getenv("POD_DELETED"), "true")

				return true
			},
		},
		{
			name:     "'happy' sad path, build cleanup command fails, but priv pod still cleaned up",
			settings: &config.KsniffSettings{},
			apiservice: ModularKubernetesApiService{
				deletePod: func(_ string) error {
					t.Setenv("POD_DELETED", "true")

					return nil
				},
			},
			bridge: ModularRuntimeBridge{
				buildCleanupCommand: func() []string {
					return nil
				},
			},
			privilegedPod: &v1.Pod{},
			expectErr:     false,
			extraAssertions: func(svc *sniffer.PrivilegedPodSnifferService) bool {
				t.Helper()

				assert.Equal(os.Getenv("POD_DELETED"), "true")

				return true
			},
		},
		{
			name:     "sad path, priv pod deletion fails",
			settings: &config.KsniffSettings{},
			apiservice: ModularKubernetesApiService{
				executeCommand: func(_, _ string, _ []string, _ io.Writer) (int, error) {
					return 0, nil
				},
				deletePod: func(_ string) error {
					return fmt.Errorf("failed to delete pod")
				},
			},
			bridge:        NopRuntimeBridge{},
			privilegedPod: &v1.Pod{},
			expectErr:     true,
			extraAssertions: func(svc *sniffer.PrivilegedPodSnifferService) bool {
				t.Helper()

				assert.NotEqual(os.Getenv("POD_DELETED"), "true")

				return true
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			svc := sniffer.NewPrivilegedPodRemoteSniffingService(
				testCase.settings,
				testCase.apiservice,
				testCase.bridge,
			)

			svc.PrivilegedPod = testCase.privilegedPod

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
