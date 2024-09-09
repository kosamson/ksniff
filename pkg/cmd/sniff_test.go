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

func TestNewCmdSniff_FlagsRegistered(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	flagKeys := []struct {
		name       string
		shorthand  string
		defaultVal string
	}{
		{"namespace", "n", ""},
		{"interface", "i", "any"},
		{"container", "c", ""},
		{"filter", "f", ""},
		{"output-file", "o", ""},
		{"local-tcpdump-path", "l", ""},
		{"remote-tcpdump-path", "r", "/tmp/static-tcpdump"},
		{"verbose", "v", "false"},
		{"privileged", "p", "false"},
		{"pod-creation-timeout", "", "1m0s"},
		{"image", "", ""},
		{"tcpdump-image", "", ""},
		{"context", "x", ""},
		{"socket", "", ""},
		{"serviceaccount", "s", ""},
	}

	cmd := cmd.NewCmdSniff(genericiooptions.NewTestIOStreamsDiscard())

	for _, flagKey := range flagKeys {
		t.Run(flagKey.name, func(t *testing.T) {
			t.Parallel()

			flag := cmd.Flags().Lookup(flagKey.name)

			assert.NotNil(flag)
			assert.Equal(flagKey.name, flag.Name)
			assert.Equal(flagKey.shorthand, flag.Shorthand)
			assert.Equal(flagKey.defaultVal, flag.DefValue)
		})
	}
}

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
