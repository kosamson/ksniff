package sniffer

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"ksniff/kube"
	"ksniff/pkg/config"
	"ksniff/pkg/service/sniffer/runtime"
)

type PrivilegedPodSnifferService struct {
	Settings                *config.KsniffSettings
	PrivilegedPod           *v1.Pod
	PrivilegedContainerName string
	TargetProcessId         *string
	KubernetesApiService    kube.KubernetesApiService
	RuntimeBridge           runtime.ContainerRuntimeBridge
}

func NewPrivilegedPodRemoteSniffingService(
	options *config.KsniffSettings,
	service kube.KubernetesApiService,
	bridge runtime.ContainerRuntimeBridge,
) *PrivilegedPodSnifferService {
	return &PrivilegedPodSnifferService{
		Settings:                options,
		PrivilegedContainerName: "ksniff-privileged",
		KubernetesApiService:    service,
		RuntimeBridge:           bridge,
	}
}

func (p *PrivilegedPodSnifferService) Setup() error {
	var err error

	log.Infof("creating privileged pod on node: '%s'", p.Settings.DetectedPodNodeName)

	if p.Settings.UseDefaultImage {
		p.Settings.Image = p.RuntimeBridge.GetDefaultImage()
	}

	if p.Settings.UseDefaultTCPDumpImage {
		p.Settings.TCPDumpImage = p.RuntimeBridge.GetDefaultTCPImage()
	}

	if p.Settings.UseDefaultSocketPath {
		p.Settings.SocketPath = p.RuntimeBridge.GetDefaultSocketPath()
	}

	p.PrivilegedPod, err = p.KubernetesApiService.CreatePrivilegedPod(
		p.Settings.DetectedPodNodeName,
		p.PrivilegedContainerName,
		p.Settings.Image,
		p.Settings.SocketPath,
		p.Settings.UserSpecifiedPodCreateTimeout,
		p.Settings.UserSpecifiedServiceAccount,
	)
	if err != nil {
		log.WithError(err).Errorf("failed to create privileged pod on node: '%s'", p.Settings.DetectedPodNodeName)
		return err
	}

	log.Infof("pod: '%s' created successfully on node: '%s'", p.PrivilegedPod.Name, p.Settings.DetectedPodNodeName)

	if p.RuntimeBridge.NeedsPid() {
		var buff bytes.Buffer

		command := p.RuntimeBridge.BuildInspectCommand(p.Settings.DetectedContainerId)

		// TODO: we need to check exit code (if non-zero) like in static tcpdump sniffer
		exitCode, err := p.KubernetesApiService.ExecuteCommand(p.PrivilegedPod.Name, p.PrivilegedContainerName, command, &buff)
		if err != nil || exitCode != 0 {
			log.WithError(err).Errorf("failed to start sniffing using privileged pod, exit code: '%d'", exitCode)

			return errors.Errorf("executing sniffer failed, exit code: '%d', error: '%s", exitCode, err)
		}

		p.TargetProcessId, err = p.RuntimeBridge.ExtractPid(buff.String())
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PrivilegedPodSnifferService) Cleanup() error {
	command := p.RuntimeBridge.BuildCleanupCommand()

	if command != nil {
		log.Infof("removing privileged container: '%s'", p.PrivilegedContainerName)

		exitCode, err := p.KubernetesApiService.ExecuteCommand(p.PrivilegedPod.Name, p.PrivilegedContainerName, command, &kube.NopWriter{})
		if err != nil {
			log.WithError(err).
				Errorf(
					"failed to remove privileged container: '%s', exit code: '%d', please manually remove it",
					p.PrivilegedContainerName,
					exitCode,
				)
		} else {
			log.Infof("privileged container: '%s' removed successfully", p.PrivilegedContainerName)
		}
	}

	if p.PrivilegedPod != nil {
		log.Infof("removing pod: '%s'", p.PrivilegedPod.Name)

		err := p.KubernetesApiService.DeletePod(p.PrivilegedPod.Name)
		if err != nil {
			log.WithError(err).Errorf("failed to remove pod: '%s", p.PrivilegedPod.Name)

			return err
		}

		log.Infof("pod: '%s' removed successfully", p.PrivilegedPod.Name)
	}

	return nil
}

func (p *PrivilegedPodSnifferService) Start(stdOut io.Writer) error {
	log.Info("starting remote sniffing using privileged pod")

	args := runtime.TcpDumpArguments{
		ContainerId:  &p.Settings.DetectedContainerId,
		NetInterface: p.Settings.UserSpecifiedInterface,
		Filter:       p.Settings.UserSpecifiedFilter,
		Pid:          p.TargetProcessId,
		SocketPath:   p.Settings.SocketPath,
		TcpdumpImage: p.Settings.TCPDumpImage,
	}

	command := p.RuntimeBridge.BuildTcpdumpCommand(args)

	exitCode, err := p.KubernetesApiService.ExecuteCommand(p.PrivilegedPod.Name, p.PrivilegedContainerName, command, stdOut)
	if err != nil {
		log.WithError(err).Errorf("failed to start sniffing using privileged pod, exit code: '%d'", exitCode)

		return err
	}

	log.Info("remote sniffing using privileged pod completed")

	return nil
}
