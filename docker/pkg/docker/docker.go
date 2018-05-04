package docker

import (
	"github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"
)

type Docker struct {
	client *dockerClient.Client
}

func (docker *Docker) ListServices() (swarmServices []swarm.Service, err error) {
	swarmService, err := docker.client.ListServices(dockerClient.ListServicesOptions{})
	if err != nil {
		log.Printf("Swarm list services failed because %v \n", err)
		return swarmService, err
	}
	log.Printf("Swarm service count: %d \n", len(swarmService))
	return swarmService, nil
}

func (docker *Docker) GetServices(id string) (swarmServices *swarm.Service, err error) {
	swarmService, err := docker.client.InspectService(id)
	log.Printf("Swarm service image name: %s \n", swarmService.Spec.TaskTemplate.ContainerSpec.Image)
	return swarmService, err
}

func (docker *Docker) GetSwarmServiceImage(swarmService swarm.Service) string {
	log.Printf("Swarm version: %v \n", swarmService.Spec.TaskTemplate.ContainerSpec.Image)
	return swarmService.Spec.TaskTemplate.ContainerSpec.Image
}

func (docker *Docker) UpdateServices(swarmService *swarm.Service, labels map[string]string) error {
	err := docker.client.UpdateService(swarmService.ID, dockerClient.UpdateServiceOptions{
		ServiceSpec: swarm.ServiceSpec{
			Annotations: swarm.Annotations{
				Labels: labels,
			},
		},
	})
	return err
}

func NewDocker() (cli *Docker, err error) {

	endpoint := "unix:///var/run/docker.sock"
	client, err := dockerClient.NewVersionedClient(endpoint, "1.24")

	return &Docker{
		client: client,
	}, err
}