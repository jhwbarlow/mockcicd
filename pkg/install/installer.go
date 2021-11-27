package install

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

type Installer interface {
	Install(imageName, imageTag string, timeout time.Duration) error
}

type HelmK8sAtomicInstaller struct {
	ReleaseName  string
	K8sNamespace string
	ChartPath    string
}

func NewHelmK8sAtomicInstaller(releaseName, k8sNamespace, chartPath string) *HelmK8sAtomicInstaller {
	return &HelmK8sAtomicInstaller{
		ReleaseName:  releaseName,
		K8sNamespace: k8sNamespace,
		ChartPath:    chartPath,
	}
}

func (i *HelmK8sAtomicInstaller) Install(imageName, imageTag string, timeout time.Duration) error {
	/*
		helm upgrade \
			--install \
			--atomic \
			--create-namespace
			--timeout "$helm_timeout" \
			-n "$k8s_namespace" \
			--set image.repository="$image_name" \
			--set image.tag="$image_tag" \
			"$release_name" \
			"$helm_chart_dir"
	*/

	log.Printf("installing Helm release %q in namespace %q", i.ReleaseName, i.K8sNamespace)

	cmd := exec.Command("helm",
		"upgrade",
		"--install",
		"--atomic",
		"--create-namespace",
		"--timeout", timeout.String(),
		"-n", i.K8sNamespace,
		"--set", "image.repository="+imageName,
		"--set", "image.tag="+imageTag,
		i.ReleaseName,
		i.ChartPath)

	log.Printf("executing command: %s", cmd.String())
	if err := cmd.Run(); err != nil {
		// TODO: Capture stderr and use in error
		return fmt.Errorf("running helm upgrade command with chart %q: %w", i.ChartPath, err)
	}
	log.Println("helm release installed")

	return nil
}
