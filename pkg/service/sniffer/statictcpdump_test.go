package sniffer_test

import (
	"fmt"
	"io"
	"testing"
	"time"

	"ksniff/kube"
	"ksniff/pkg/config"
	"ksniff/pkg/service/sniffer"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	NopKubernetesApiService     struct{}
	ErrKubernetesApiService     struct{}
	ModularKubernetesApiService struct {
		executeCommand      func(string, string, []string, io.Writer) (int, error)
		deletePod           func(string) error
		createPrivilegedPod func(string, string, string, string, time.Duration, string) (*corev1.Pod, error)
		uploadFile          func(string, string, string, string) error
	}
)

// ensure mocked services implement interface
var (
	_ kube.KubernetesApiService = NopKubernetesApiService{}
	_ kube.KubernetesApiService = ErrKubernetesApiService{}
	_ kube.KubernetesApiService = ErrKubernetesApiService{}
)

func (svc NopKubernetesApiService) ExecuteCommand(_, _ string, _ []string, _ io.Writer) (int, error) {
	return 0, nil
}

func (svc NopKubernetesApiService) DeletePod(_ string) error {
	return nil
}

func (svc NopKubernetesApiService) CreatePrivilegedPod(_, _, _, _ string, _ time.Duration, _ string) (*corev1.Pod, error) {
	return &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: "custom-pod",
		},
	}, nil
}

func (svc NopKubernetesApiService) UploadFile(_, _, _, _ string) error {
	return nil
}

func (svc ErrKubernetesApiService) ExecuteCommand(_, _ string, _ []string, _ io.Writer) (int, error) {
	return -1, fmt.Errorf("")
}

func (svc ErrKubernetesApiService) DeletePod(_ string) error {
	return fmt.Errorf("")
}

func (svc ErrKubernetesApiService) CreatePrivilegedPod(_, _, _, _ string, _ time.Duration, _ string) (*corev1.Pod, error) {
	return nil, fmt.Errorf("")
}

func (svc ErrKubernetesApiService) UploadFile(_, _, _, _ string) error {
	return fmt.Errorf("")
}

func (svc ModularKubernetesApiService) ExecuteCommand(podName, containerName string, command []string, stdOut io.Writer) (int, error) {
	return svc.executeCommand(podName, containerName, command, stdOut)
}

func (svc ModularKubernetesApiService) DeletePod(podName string) error {
	return svc.deletePod(podName)
}

func (svc ModularKubernetesApiService) CreatePrivilegedPod(nodeName, containerName, image, socketPath string, timeout time.Duration, serviceaccount string) (*corev1.Pod, error) {
	return svc.createPrivilegedPod(nodeName, containerName, image, socketPath, timeout, serviceaccount)
}

func (svc ModularKubernetesApiService) UploadFile(localPath, remotePath, podName, containerName string) error {
	return svc.uploadFile(localPath, remotePath, podName, containerName)
}

func TestSetup_Static(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	testCases := []struct {
		name       string
		settings   *config.KsniffSettings
		apiservice kube.KubernetesApiService
		expectErr  bool
	}{
		{
			name:       "happy path",
			settings:   &config.KsniffSettings{},
			apiservice: NopKubernetesApiService{},
			expectErr:  false,
		},
		{
			name:       "sad path",
			settings:   &config.KsniffSettings{},
			apiservice: ErrKubernetesApiService{},
			expectErr:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			svc := sniffer.NewUploadTcpdumpRemoteSniffingService(
				testCase.settings,
				testCase.apiservice,
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

func TestStart_Static(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	testCases := []struct {
		name       string
		settings   *config.KsniffSettings
		apiservice kube.KubernetesApiService
		expectErr  bool
	}{
		{
			name:       "happy path",
			settings:   &config.KsniffSettings{},
			apiservice: NopKubernetesApiService{},
			expectErr:  false,
		},
		{
			name:     "sad path: error returned",
			settings: &config.KsniffSettings{},
			apiservice: ModularKubernetesApiService{
				executeCommand: func(s1, s2 string, s3 []string, w io.Writer) (int, error) {
					return 0, fmt.Errorf("")
				},
			},
			expectErr: true,
		},
		{
			name:     "sad path: unexpected error code",
			settings: &config.KsniffSettings{},
			apiservice: ModularKubernetesApiService{
				executeCommand: func(s1, s2 string, s3 []string, w io.Writer) (int, error) {
					return 1, nil
				},
			},
			expectErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			svc := sniffer.NewUploadTcpdumpRemoteSniffingService(
				testCase.settings,
				testCase.apiservice,
			)

			err := svc.Start(nil)

			if testCase.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

// this test is pretty useless until actual cleanup logic happens
func TestCleanup_Static(t *testing.T) {
	t.Parallel()

	staticSniffer := sniffer.NewUploadTcpdumpRemoteSniffingService(nil, nil)

	assert.NoError(t, staticSniffer.Cleanup())
}
