package sniffer

import (
	"io"
	"ksniff/kube"
	"ksniff/pkg/config"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	EPHEMERAL_CONTAINER_NAME = "ksniff-ephemeral"
)

type EphemeralContainerSnifferService struct {
	settings             *config.KsniffSettings
	kubernetesApiService kube.KubernetesApiService
}

func NewEphemeralContainerSniffingService(options *config.KsniffSettings, service kube.KubernetesApiService) SnifferService {
	return &EphemeralContainerSnifferService{settings: options, kubernetesApiService: service}
}

func (u *EphemeralContainerSnifferService) Setup() error {
	log.Infof("uploading static tcpdump binary from: '%s' to: '%s'",
		u.settings.UserSpecifiedLocalTcpdumpPath, u.settings.UserSpecifiedRemoteTcpdumpPath)

	_, err := u.kubernetesApiService.CreateEphemeralContainer(u.settings.UserSpecifiedPodName)

	if err != nil {
		panic(err)
	}

	err = u.kubernetesApiService.UploadFile(u.settings.UserSpecifiedLocalTcpdumpPath,
		u.settings.UserSpecifiedRemoteTcpdumpPath, u.settings.UserSpecifiedPodName, EPHEMERAL_CONTAINER_NAME)

	if err != nil {
		log.WithError(err).Errorf("failed uploading static tcpdump binary to ephemeral container, please verify the remote container has tar installed")
		return err
	}

	log.Info("tcpdump uploaded successfully")

	return nil
}

func (u *EphemeralContainerSnifferService) Cleanup() error {
	return nil
}

func (u *EphemeralContainerSnifferService) Start(stdOut io.Writer) error {
	log.Info("start sniffing on ephemeral container")

	command := []string{u.settings.UserSpecifiedRemoteTcpdumpPath, "-i", u.settings.UserSpecifiedInterface,
		"-U", "-w", "-", u.settings.UserSpecifiedFilter}

	exitCode, err := u.kubernetesApiService.ExecuteCommand(u.settings.UserSpecifiedPodName, EPHEMERAL_CONTAINER_NAME, command, stdOut)
	if err != nil || exitCode != 0 {
		return errors.Errorf("executing sniffer failed, exit code: '%d'", exitCode)
	}

	log.Infof("done sniffing on ephemeral container")

	return nil
}
