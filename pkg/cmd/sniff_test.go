package cmd_test

import (
	"fmt"
	"strings"
	"testing"

	"ksniff/pkg/cmd"
	"ksniff/pkg/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestNewCmdSniff_FlagsRegistered(t *testing.T) {
	assert := assert.New(t)

	// TODO: should test if pflag is properly bound to
	// ksniffsettings struct fields
	flagKeys := []struct {
		name       string
		shorthand  string
		defaultVal string
		envKey     string
	}{
		{"namespace", "n", "", "KUBECTL_PLUGINS_CURRENT_NAMESPACE"},
		{"interface", "i", "any", "KUBECTL_PLUGINS_LOCAL_FLAG_INTERFACE"},
		{"container", "c", "", "KUBECTL_PLUGINS_LOCAL_FLAG_CONTAINER"},
		{"filter", "f", "", "KUBECTL_PLUGINS_LOCAL_FLAG_FILTER"},
		{"output-file", "o", "", "KUBECTL_PLUGINS_LOCAL_FLAG_OUTPUT_FILE"},
		{"local-tcpdump-path", "l", "", "KUBECTL_PLUGINS_LOCAL_FLAG_LOCAL_TCPDUMP_PATH"},
		{"remote-tcpdump-path", "r", "/tmp/static-tcpdump", "KUBECTL_PLUGINS_LOCAL_FLAG_REMOTE_TCPDUMP_PATH"},
		{"verbose", "v", "false", "KUBECTL_PLUGINS_LOCAL_FLAG_VERBOSE"},
		{"privileged", "p", "false", "KUBECTL_PLUGINS_LOCAL_FLAG_PRIVILEGED"},
		{"pod-creation-timeout", "", "1m0s", ""},
		{"image", "", "", "KUBECTL_PLUGINS_LOCAL_FLAG_IMAGE"},
		{"tcpdump-image", "", "", "KUBECTL_PLUGINS_LOCAL_FLAG_TCPDUMP_IMAGE"},
		{"context", "x", "", "KUBECTL_PLUGINS_CURRENT_CONTEXT"},
		{"socket", "", "", "KUBECTL_PLUGINS_SOCKET_PATH"},
		{"serviceaccount", "s", "", "KUBECTL_PLUGINS_LOCAL_FLAG_SERVICE_ACCOUNT"},
	}

	cmd := cmd.NewCmdSniff(genericiooptions.NewTestIOStreamsDiscard())
	boundViperKeys := stringSliceToSet(viper.AllKeys())

	for _, flagKey := range flagKeys {
		t.Run(flagKey.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(flagKey.name)

			assert.NotNil(flag)
			assert.Equal(flagKey.name, flag.Name)
			assert.Equal(flagKey.shorthand, flag.Shorthand)
			assert.Equal(flagKey.defaultVal, flag.DefValue)

			// check if viper keys bound to environment variables
			if flagKey.envKey != "" {
				envVal := "test-env-" + flagKey.name

				t.Setenv(flagKey.envKey, envVal)
				assert.Equal(envVal, viper.GetString(flagKey.name))
			}

			// check if viper keys bound to cobra pflags
			//
			// this logic is intentionally placed *after*
			// testing environment variable binding
			// since flags take a higher precedence in
			// viper config management
			if _, ok := boundViperKeys[flagKey.name]; ok {
				var flagVal string
				if flagKey.defaultVal != "" {
					flagVal = flagKey.defaultVal
				} else {
					flagVal = "test-flag-" + flagKey.name
				}

				flagArg := fmt.Sprintf("--%s=%s", flagKey.name, flagVal)

				cmd.ParseFlags([]string{flagArg}) // why does cmd.SetArgs() not work?
				assert.Equal(flagVal, viper.GetString(flagKey.name))
			}
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

func TestComplete_VerboseModeSpecified(t *testing.T) {
	t.Parallel()

	settings := config.NewKsniffSettings(genericiooptions.IOStreams{})

	sniff := cmd.NewKsniff(settings)
	cmd := cmd.NewCmdSniff(genericclioptions.IOStreams{})

	cmd.ParseFlags([]string{"--verbose"})

	var commands []string

	err := sniff.Complete(cmd, append(commands, "pod-name"))

	assert.Nil(t, err)
	assert.Equal(t, log.DebugLevel, log.GetLevel())
}

func TestComplete_UserSpecifiedKubeContextNotExists(t *testing.T) {
	t.Parallel()

	settings := config.NewKsniffSettings(genericiooptions.IOStreams{})

	sniff := cmd.NewKsniff(settings)
	cmd := cmd.NewCmdSniff(genericclioptions.IOStreams{})

	cmd.ParseFlags([]string{"--context=fake-context"})

	var commands []string

	err := sniff.Complete(cmd, append(commands, "pod-name"))

	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "context doesn't exist")
}

func stringSliceToSet(sl []string) map[string]bool {
	mp := make(map[string]bool, len(sl))

	for _, s := range sl {
		mp[s] = true
	}

	return mp
}
