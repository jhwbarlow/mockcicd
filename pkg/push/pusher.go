package push

import (
	"fmt"
	"log"
	"os/exec"
)

type Pusher interface {
	Push(name, tag string) error
}

type DockerCLIPusher struct{}

func NewDockerCLIPusher() *DockerCLIPusher {
	return new(DockerCLIPusher)
}

func (*DockerCLIPusher) Push(name, tag string) error {
	/*
		docker push "${image_name}:${image_tag}"
	*/

	fullImageName := name + ":" + tag
	log.Printf("pushing docker image %q", fullImageName)

	cmd := exec.Command("docker", "push", fullImageName)

	log.Printf("executing command: %s", cmd.String())
	if err := cmd.Run(); err != nil {
		// TODO: Capture stderr and use in error
		return fmt.Errorf("running docker push command for image %q: %w", fullImageName, err)
	}
	log.Println("image pushed")

	return nil
}
