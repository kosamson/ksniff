package runtime_test

import (
	"testing"

	"ksniff/pkg/service/sniffer/runtime"

	"github.com/stretchr/testify/assert"
)

func TestNewContainerRuntimeBridge_Docker(t *testing.T) {
	bridge := runtime.NewContainerRuntimeBridge("docker")
	assert.IsType(t, &runtime.DockerBridge{}, bridge)
}

func TestNewContainerRuntimeBridge_Containerd(t *testing.T) {
	bridge := runtime.NewContainerRuntimeBridge("containerd")
	assert.IsType(t, &runtime.ContainerdBridge{}, bridge)
}

func TestNewContainerRuntimeBridge_Crio(t *testing.T) {
	bridge := runtime.NewContainerRuntimeBridge("cri-o")
	assert.IsType(t, &runtime.CrioBridge{}, bridge)
}

func TestNewContainerRuntimeBridge_Invalid(t *testing.T) {
	assert.Panics(t, func() { runtime.NewContainerRuntimeBridge("i-do-not-exist") })
}

// helper for other runtime test files
func stringPtr(s string) *string {
	return &s
}
