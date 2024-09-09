package cmd_test

import (
	"strings"
	"testing"

	"ksniff/pkg/cmd"
	"ksniff/pkg/config"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestComplete_NotEnoughArguments(t *testing.T) {
	t.Parallel()

	// given
	settings := config.NewKsniffSettings(genericiooptions.IOStreams{})
	sniff := cmd.NewKsniff(settings)
	cmd := &cobra.Command{}
	var commands []string

	// when
	err := sniff.Complete(cmd, commands)

	// then
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "not enough arguments"))
}

func TestComplete_EmptyPodName(t *testing.T) {
	t.Parallel()

	// given
	settings := config.NewKsniffSettings(genericiooptions.IOStreams{})
	sniff := cmd.NewKsniff(settings)
	cmd := &cobra.Command{}
	var commands []string

	// when
	err := sniff.Complete(cmd, append(commands, ""))

	// then
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "pod name is empty"))
}

func TestComplete_PodNameSpecified(t *testing.T) {
	t.Parallel()

	// given
	settings := config.NewKsniffSettings(genericiooptions.IOStreams{})
	sniff := cmd.NewKsniff(settings)
	cmd := cmd.NewCmdSniff(genericclioptions.IOStreams{})
	var commands []string

	// when
	err := sniff.Complete(cmd, append(commands, "pod-name"))

	// then
	assert.Nil(t, err)
	assert.Equal(t, "pod-name", settings.UserSpecifiedPodName)
}
