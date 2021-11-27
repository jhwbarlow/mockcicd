package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/jhwbarlow/mockcicd/pkg/build"
	"github.com/jhwbarlow/mockcicd/pkg/check"
	"github.com/jhwbarlow/mockcicd/pkg/install"
	"github.com/jhwbarlow/mockcicd/pkg/obtain"
	"github.com/jhwbarlow/mockcicd/pkg/prepare"
	"github.com/jhwbarlow/mockcicd/pkg/push"
	"github.com/jhwbarlow/mockcicd/pkg/tagdeduce"
)

type config struct {
	SrcDirPath       string        `required:"true"`
	GitRepoURL       string        `required:"true"`
	GitBranch        string        `required:"true"`
	ImageName        string        `required:"true"`
	HelmChartPath    string        `required:"true"`
	HelmK8sNamespace string        `required:"true"`
	HelmReleaseName  string        `required:"true"`
	InstallTimeout   time.Duration `required:"true"`
	PollPeriod       time.Duration `required:"true"`
}

const (
	appName = "mockcicd"
)

func main() {
	// Get config from environment.
	// TODO: This can be abstracted into an interface and unit tested.
	config := new(config)
	if err := envconfig.Process(appName, config); err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// Create dependencies for injection
	preparer := prepare.NewFilesystemPreparer()
	obtainer := obtain.NewGitCloneObtainer(config.GitRepoURL, config.GitBranch, preparer)
	tagDeducer := tagdeduce.NewGitHashTagDeducer(config.SrcDirPath)
	builder := build.NewDockerCLIBuilder()
	pusher := push.NewDockerCLIPusher()
	installer := install.NewHelmK8sAtomicInstaller(config.HelmReleaseName,
		config.HelmK8sNamespace,
		config.HelmChartPath)
	checker := check.NewGitChecker(config.SrcDirPath, config.GitBranch)
	updater := obtain.NewGitPullUpdater(config.GitBranch)

	if err := setup(obtainer,
		tagDeducer,
		builder,
		pusher,
		installer,
		config.SrcDirPath,
		config.ImageName,
		config.InstallTimeout); err != nil {
		log.Fatalf("Setup error: %v", err)
	}

	// Run does not return
	run(tagDeducer,
		builder,
		pusher,
		installer,
		checker,
		updater,
		config.SrcDirPath,
		config.ImageName,
		config.PollPeriod,
		config.InstallTimeout,
		nil)
}

func setup(obtainer obtain.Obtainer,
	tagDeducer tagdeduce.TagDeducer,
	builder build.Builder,
	pusher push.Pusher,
	installer install.Installer,
	srcDirPath string,
	imageName string,
	installTimeout time.Duration) error {
	if err := obtainer.Obtain(srcDirPath); err != nil {
		return fmt.Errorf("obtaining source code: %w", err)
	}

	// Initial build and installation
	if err := buildAndInstall(tagDeducer,
		builder,
		pusher,
		installer,
		srcDirPath,
		imageName,
		installTimeout); err != nil {
		return fmt.Errorf("performing initial build and install: %w", err)
	}

	return nil
}

func run(tagDeducer tagdeduce.TagDeducer,
	builder build.Builder,
	pusher push.Pusher,
	installer install.Installer,
	checker check.Checker,
	updater obtain.Updater,
	srcDirPath string,
	imageName string,
	pollPeriod time.Duration,
	installTimeout time.Duration,
	done <-chan struct{}) {
	// Check for changes, build and install them.
	// If the done channel is closed, stop polling and return.
	for {
		select {
		case <-done:
			return
		default:
		}

		time.Sleep(pollPeriod)

		hasChanged, err := checker.Check()
		if err != nil {
			// If there is an error, try again next time
			log.Printf("Warning: Error checking for changes: %v", err)
			continue
		}

		if hasChanged {
			if err := updater.Update(srcDirPath); err != nil {
				// If there is an error, try again next time
				log.Printf("Warning: Error updating to obtain latest changes: %v", err)
				continue
			}

			if err := buildAndInstall(tagDeducer,
				builder,
				pusher,
				installer,
				srcDirPath,
				imageName,
				installTimeout); err != nil {
				// If there is an error, the installer must deal with it and leave the app in a
				// working state. Therefore, we await the next change which may fix the error.
				log.Printf("Warning: Error performing build and install due to change: %v", err)
				continue
			}
		}
	}
}

func buildAndInstall(tagDeducer tagdeduce.TagDeducer,
	builder build.Builder,
	pusher push.Pusher,
	installer install.Installer,
	srcDirPath string,
	imageName string,
	installTimeout time.Duration) error {

	tag, err := tagDeducer.Deduce()
	if err != nil {
		return fmt.Errorf("deducing tag: %w", err)
	}

	if err := builder.Build(srcDirPath, imageName, tag); err != nil {
		return fmt.Errorf("building image: %w", err)
	}

	if err := pusher.Push(imageName, tag); err != nil {
		return fmt.Errorf("pushing image: %w", err)
	}

	if err := installer.Install(imageName, tag, installTimeout); err != nil {
		return fmt.Errorf("installing image: %w", err)
	}

	return nil
}
