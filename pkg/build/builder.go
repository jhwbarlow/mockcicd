package build

import (
	"fmt"
	"log"
	"os/exec"
)

type Builder interface {
	Build(buildContextPath, name, tag string) error
}

type DockerCLIBuilder struct{}

func NewDockerCLIBuilder() *DockerCLIBuilder {
	return new(DockerCLIBuilder)
}

func (*DockerCLIBuilder) Build(buildContextPath, name, tag string) error {
	/*
		docker build \
		-f docker/Dockerfile \
		-t "${image_name}:${image_tag}" \
		"$src_dir"
	*/

	fullImageName := name + ":" + tag
	log.Printf("building docker image %q", fullImageName)

	cmd := exec.Command("docker",
		"build",
		"-f", "docker/Dockerfile",
		"-t", fullImageName,
		buildContextPath)

	log.Printf("executing command: %s", cmd.String())
	if err := cmd.Run(); err != nil {
		// TODO: Capture stderr and use in error
		return fmt.Errorf("running docker build command with context %q: %w", buildContextPath, err)
	}
	log.Println("image built")

	return nil
}
