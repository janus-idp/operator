## End-to-end tests

The end-to-end tests use the [Ginkgo framework](https://onsi.github.io/ginkgo/) and allow to test the operator against a real cluster in the following scenarios:
- building and deploying the operator image off of the current code
- using a specific image or a specific downstream build

Deployment of the operator itself can be done by:
- deploying with or without OLM,
- or deploying the downstream bundle in both online and air-gapped scenarios

To run the end-to-end tests, you can use:
```shell
$ make test-e2e
```

### Configuration

The behavior is configurable using the following environment variables:

| Name                                                                                           | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                   | Default value                                     | Example                                                 |
|------------------------------------------------------------------------------------------------|--------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------|---------------------------------------------------------|
| `BACKSTAGE_OPERATOR_TEST_MODE`                                                                 | string | The test mode:<br>- if not set, it will call `make deploy`<br>- `olm`: it will call `make deploy-olm`<br>- `rhdh-latest` or `rhdh-next`: it will install the operator using the [`install-rhdh-catalog-source.sh`](../../.rhdh/scripts/install-rhdh-catalog-source.sh) script<br>- `rhdh-airgap`: it will install the operator using the [`prepare-restricted-environment.sh`](../../.rhdh/scripts/prepare-restricted-environment.sh) script. |                                                   | `rhdh-latest`                                           |
| `IMG` (or any variables from the Makefile that are used by `make deploy` or `make deploy-olm`) | string | The image to use. Relevant if `BACKSTAGE_OPERATOR_TEST_MODE` is not set or set to `olm`                                                                                                                                                                                                                                                                                                                                                       | `VERSION` defined in [`Makefile`](../../Makefile) | `quay.io/janus-idp/operator:0.0.1-latest`               |
| `BACKSTAGE_OPERATOR_TESTS_BUILD_IMAGES`                                                        | bool   | If set to `true`, it will build the operator image with `make image-build`.<br>Relevant if `BACKSTAGE_OPERATOR_TEST_MODE` is not set or set to `olm`.                                                                                                                                                                                                                                                                                         |                                                   | `false`                                                 |
| `BACKSTAGE_OPERATOR_TESTS_PUSH_IMAGES`                                                         | bool   | If set to `true`, it will push the operator image with `make image-push`.<br>Relevant if `BACKSTAGE_OPERATOR_TEST_MODE` is not set or set to `olm`.                                                                                                                                                                                                                                                                                           |                                                   | `false`                                                 |
| `BACKSTAGE_OPERATOR_TESTS_PLATFORM`                                                            | string | The platform type, to directly load the operator image if supported instead of pushing it.<br>Relevant if `BACKSTAGE_OPERATOR_TEST_MODE` is not set or set to `olm`.br>Supported values: [`kind`](#building-and-testing-local-changes-on-kind), [`k3d`](#building-and-testing-local-changes-on-k3d), [`minikube`](#building-and-testing-local-changes-on-minikube)                                                                            |                                                   | `kind`                                                  |
| `BACKSTAGE_OPERATOR_TESTS_KIND_CLUSTER`                                                        | string | Name of the local KinD cluster to use. Relevant only if `BACKSTAGE_OPERATOR_TESTS_PLATFORM` is `kind`.                                                                                                                                                                                                                                                                                                                                        | `kind`                                            | `kind-local-k8s-cluster`                                |
| `BACKSTAGE_OPERATOR_TESTS_K3D_CLUSTER`                                                         | string | Name of the local k3d cluster to use. Relevant only if `BACKSTAGE_OPERATOR_TESTS_PLATFORM` is `k3d`.                                                                                                                                                                                                                                                                                                                                          | `k3s-default`                                     | `k3d-local-k8s-cluster`                                 |
| `BACKSTAGE_OPERATOR_TESTS_AIRGAP_INDEX_IMAGE`                                                  | string | Index image to use in the airgap scenario.<br>Relevant if `BACKSTAGE_OPERATOR_TEST_MODE` is `rhdh-airgap`.                                                                                                                                                                                                                                                                                                                                    | `quay.io/rhdh/iib:latest-v4.14-x86_64`            | `registry.redhat.io/redhat/redhat-operator-index:v4.14` |
| `BACKSTAGE_OPERATOR_TESTS_AIRGAP_OPERATOR_VERSION`                                             | string | Operator version to use in the airgap scenario.<br>Relevant if `BACKSTAGE_OPERATOR_TEST_MODE` is `rhdh-airgap`.                                                                                                                                                                                                                                                                                                                               | `v1.1.0`                                          | `v1.1.0`                                                |
| `BACKSTAGE_OPERATOR_TESTS_AIRGAP_MIRROR_REGISTRY`                                              | string | Existing mirror registry to use in the airgap scenario.<br>Relevant if `BACKSTAGE_OPERATOR_TEST_MODE` is `rhdh-airgap`<br>.                                                                                                                                                                                                                                                                                                                   |                                                   | `my-registry.example.com`                               |

### Examples

#### Testing a specific version

This should work on any Kubernetes cluster:

```shell
$ make test-e2e VERSION=0.0.1-latest
```

#### Building and testing local changes on [kind](https://kind.sigs.k8s.io/)

```shell
$ kind create cluster
$ make test-e2e BACKSTAGE_OPERATOR_TESTS_BUILD_IMAGES=true BACKSTAGE_OPERATOR_TESTS_PLATFORM=kind
```

#### Building and testing local changes on [k3d](https://k3d.io/)

```shell
$ k3d cluster create
$ make test-e2e BACKSTAGE_OPERATOR_TESTS_BUILD_IMAGES=true BACKSTAGE_OPERATOR_TESTS_PLATFORM=k3d
```

#### Building and testing local changes on [minikube](https://minikube.sigs.k8s.io/docs/)

```shell
$ minikube start
$ make test-e2e BACKSTAGE_OPERATOR_TESTS_BUILD_IMAGES=true BACKSTAGE_OPERATOR_TESTS_PLATFORM=minikube
```

#### Testing a specific image (e.g. PR image)

```shell
$ make test-e2e IMG=quay.io/janus-idp/operator:0.0.1-pr-201-7d08c24
```

#### Testing a specific version using OLM

This requires the [Operator Lifecycle Manager (OLM)](https://olm.operatorframework.io/) to be installed in the cluster:

```shell
$ make test-e2e BACKSTAGE_OPERATOR_TEST_MODE=olm
```

#### Testing a downstream build of RHDH

This requires an OpenShift cluster. If testing a CI build, please follow the instructions in [Installing CI builds of Red Hat Developer Hub](../../.rhdh/docs/installing-ci-builds.adoc) to add your Quay token to the cluster.

```shell
# latest
$ make test-e2e BACKSTAGE_OPERATOR_TEST_MODE=rhdh-latest

# or next
$ make test-e2e BACKSTAGE_OPERATOR_TEST_MODE=rhdh-next
```

#### Airgap testing of RHDH

This requires an OpenShift cluster.
Please also read the prerequisites in [Installing Red Hat Developer Hub (RHDH) in restricted environments](../../.rhdh/docs/airgap.adoc).

```shell
$ make test-e2e BACKSTAGE_OPERATOR_TEST_MODE=rhdh-airgap
```
