"Mock" CI-CD for Algolia Search Demo
====================================

This project uses Kubernetes and Helm to deploy the Algolia Search Demo.

Using the inbuilt readiness probe and the default Deployment rollout functionality of Kubernetes, it ensures that the application remains available even if a new release is faulty or an error occurs during an update.

Helm is used to automate the process of deploying the new releases into Kubernetes. The "atomic" functionality is used so that new releases which are not ready within a configurable timeout are automatically cancelled and the Deployment continues using the previous release version.

In order to "drive" the process of building a new release from source and deploying to Kubernetes, a "mock" CI-CD process is provided which performs the following steps:

1. Polls, at a configurable frequency, for new commits to a branch on a Git repository.
2. Uses the commit hash of the latest commit as the container image tag.
3. Packages the source code into a container image using Docker, tagging it with the tag from step (2).
4. Pushes the container image to a container registry using Docker.
5. Deploys the container image to Kubernetes using a Helm chart.

Implementation
--------------

The implementation of the mock CI-CD process is in Go. The structure is as follows:

- At startup, any existing source code is removed and a fresh clone of the Git repository is made. This is then used to build and push a container image, which is then deployed using Helm. If this initial deployment should fail at any step, the process will exit with an error.
- After the initial deployment is successful and the application is running, the Git repository is frequently polled, and by comparing the current local head commit hash with the remote head commit hash, decides if a new version has been pushed to the remote repository.
- If a new version has been released, the same steps as at startup are performed: The image is built, pushed, and deployed.

A proof-of-concept shell script (`mock-ci-cd`) is also provided which performs the same steps as the Go program.

Code Architecture
-----------------

The `main` package contains the main "driving" logic of the program, loading the configuration from the environment, performing the initial setup and initial deployment, and then looping forever performing the main reconciliation loop to install new releases.

The `check` package contains the functionality which checks if a new release is available by comparing Git hashes.

The `build` package contains the functionality which builds a new container image by using the external `docker` CLI binary.

The `push` package contains the functionality which pushes the container image to a registry by using the external `docker` CLI binary.

The `obtain` package contains the functionality which both initialises the local repository by cloning the remote, and also updates the local repository by pulling. This uses a package implementing the Git protocol rather than by using the external `git` binary.

The `prepare` package contains the functionality which prepares the local filesystem for cloning the remote, and is used by the the `obtain` package.

The `tagdeduce` package contains the functionality which is used to deduce what tag to assign to a container image. The current implementation uses the Git hash of the latest commit.

The `install` package contains the functionality which installs the new release. The current implementation installs by deploying to Kubernetes using the `helm` CLI binary.

The `git` package contains utility routines which interact with the Git repositories, and is used by the other packages.

Requirements
------------

The following requirements are assumed for the Go program:

- `docker` binary, with the running user having permissions to call the Docker daemon, and logged in to the registry to which images will be pushed.
- `helm` version 3 binary, with the user having a appropriate default `kubeconfig` so that Helm can access the chosen Kubernetes cluster (e.g. a local Minikube).

Running
-------

A compiled version of the Go program is not provided. Instead, the Go can be run as a "script" using `go run main.go` The `run-go.sh` is a wrapper around this and provides the configuration environment variables.

Configuration
-------------

The following environment variables configure the Go program:
- *MOCKCICD_SRCDIRPATH* - the path where the source code will be stored. (Note: Only tested as a path relative to the root of this project).
- *MOCKCICD_GITREPOURL* - the URL of the Git repository containing the source code to deploy. Currently only supports HTTPS Git URLs, not SSH.
- *MOCKCICD_GITBRANCH* - the branch in the repository.
- *MOCKCICD_IMAGENAME* - the name of the container image (including the registry name) that will be built and pushed.
- *MOCKCICD_HELMCHARTPATH* - the path to the Helm chart to be used. (Note: Only tested using the provided chart path relative to the root of this project).
- *MOCKCICD_HELMK8SNAMESPACE* - the Kubernetes namespace to which the application will be deployed.
- *MOCKCICD_HELMRELEASENAME* - the Helm release name that will be used.
- *MOCKCICD_INSTALLTIMEOUT* - the install timeout. If a new release is not ready by this time, it will be automatically rolled-back.
- *MOCKCICD_POLLPERIOD* - the period between checking the Git repository for changes indicating new releases. 

Unit Tests
----------

As the majority of the code, by its very nature, deals with communicating with external systems, this presents some challenges for unit testing. However, there are comprehensive unit tests provided for the `main` package. These test the main logic of the startup and reconciliation routines, such as ensuring that the reconciliation process continues when different types of failure are encountered.

These tests are in `main_test.go`.

Notes
-----

- The chart contains an `Ingress` resource which is bound to the `/` path. if this interferes with other ingresses, it can be changed or disabled by setting `ingress.enabled` to `false` in the Chart's `values.yaml` file.

