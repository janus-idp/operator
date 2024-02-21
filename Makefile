# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.2.0

# Using docker or podman to build and push images
CONTAINER_ENGINE ?= docker

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# quay.io/janus-idp/operator-bundle:$VERSION and quay.io/janus-idp/operator-catalog:$VERSION.
IMAGE_TAG_BASE ?= quay.io/janus-idp/operator

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:$(VERSION)

# BUNDLE_GEN_FLAGS are the flags passed to the operator-sdk generate bundle command
BUNDLE_GEN_FLAGS ?= -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)

# USE_IMAGE_DIGESTS defines if images are resolved via tags or digests
# You can enable this value if you would like to use SHA Based Digests
# To enable set flag to true
USE_IMAGE_DIGESTS ?= false
ifeq ($(USE_IMAGE_DIGESTS), true)
	BUNDLE_GEN_FLAGS += --use-image-digests
endif

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_TAG_BASE):$(VERSION)
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.25.0

# Default Backstage config directory to use
# it has to be defined as a set of YAML files inside ./config/manager/$(CONF_DIR) directory
# to use other config - add a directory with config and run 'CONF_DIR=<dir-name> make ...'
# TODO find better place than ./config/manager (but not ./config/overlays) ?
# TODO it works only for make run, needs supporting make deploy as well https://github.com/janus-idp/operator/issues/47
CONF_DIR ?= default-config

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: goimports ## Format the code using goimports
	find . -not -path '*/\.*' -name '*.go' -exec $(GOIMPORTS) -w {} \;

fmt_license: addlicense ## Ensure the license header is set on all files
	$(ADDLICENSE) -v -f license_header.txt $$(find . -not -path '*/\.*' -name '*.go')
	$(ADDLICENSE) -v -f license_header.txt $$(find . -name '*ockerfile')

.PHONY: gosec
gosec: addgosec ## run the gosec scanner for non-test files in this repo
  	# we let the report content trigger a failure using the GitHub Security features.
	$(GOSEC) -no-fail -fmt $(GOSEC_FMT) -out $(GOSEC_OUTPUT_FILE)  ./...

.PHONY: lint
lint: golangci-lint ## Run the linter on the codebase
	$(GOLANGCI_LINT) run ./... --timeout 15m

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests. We need LOCALBIN=$(LOCALBIN) to get correct default-config path
	mkdir -p $(LOCALBIN)/default-config && cp config/manager/$(CONF_DIR)/* $(LOCALBIN)/default-config
	LOCALBIN=$(LOCALBIN) KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet build ## Run a controller from your host.
	cd $(LOCALBIN) && mkdir -p default-config && cp ../config/manager/$(CONF_DIR)/* default-config && ./manager

# by default images expire from quay registry after 14 days
# set a longer timeout (or set no label to keep images forever)
LABEL ?= quay.expires-after=14d
PLATFORM ?= linux/amd64
# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64 ). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/

.PHONY: image-build
image-build: test ## Build docker image with the manager using docker.
	$(CONTAINER_ENGINE) build --platform $(PLATFORM) -f docker/Dockerfile -t $(IMG) --label $(LABEL) .

.PHONY: image-push
image-push: ## Push IMG image to registry
	$(CONTAINER_ENGINE) push $(IMG)

# PLATFORMS defines the target platforms for  the manager image be build to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enable BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG=<myregistry/image:tag> than the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
# If more arches are needed, use comma-separated list: linux/amd64,linux/arm64,linux390x,linux/ppc64le
PLATFORMS ?= linux/amd64
.PHONY: docker-buildx
docker-buildx: test ## Build and push docker image for the manager for cross-platform support
	# see https://docs.docker.com/build/building/multi-platform/#cross-compilation for BUILDPLATFORM definition
	# copy existing docker/Dockerfile and insert --platform=${BUILDPLATFORM} into docker/Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' docker/Dockerfile > docker/Dockerfile.cross
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- docker buildx build --push --platform $(PLATFORMS) --tag $(IMG) . -f docker/Dockerfile.cross  --label $(LABEL)
	- docker buildx rm project-v3-builder
	rm docker/Dockerfile.cross

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GOIMPORTS ?= $(LOCALBIN)/goimports
ADDLICENSE ?= $(LOCALBIN)/addlicense
GOSEC ?= $(LOCALBIN)/gosec

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.11.3
GOLANGCI_LINT_VERSION ?= v1.55.2
GOIMPORTS_VERSION ?= v0.15.0
ADDLICENSE_VERSION ?= v1.1.1
# opm and operator-sdk version
OPM_VERSION ?= v1.36.0
OPERATOR_SDK_VERSION ?= v1.33.0
GOSEC_VERSION ?= v2.18.2

## Gosec options - default format is sarif so we can integrate with Github code scanning
GOSEC_FMT ?= sarif  # for other options, see https://github.com/securego/gosec#output-formats
GOSEC_OUTPUT_FILE ?= gosec.sarif

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	test -s $(LOCALBIN)/golangci-lint || GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: goimports
goimports: $(GOIMPORTS) ## Download goimports locally if necessary.
$(GOIMPORTS): $(LOCALBIN)
	test -s $(LOCALBIN)/goimports || GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)

.PHONY: addlicense
addlicense: $(ADDLICENSE) ## Download addlicense locally if necessary.
$(ADDLICENSE): $(LOCALBIN)
	test -s $(LOCALBIN)/addlicense || GOBIN=$(LOCALBIN) go install github.com/google/addlicense@$(ADDLICENSE_VERSION)

.PHONY: addgosec
addgosec: $(GOSEC) ## Download gosec locally if necessary.
$(GOSEC): $(LOCALBIN)
	test -s $(LOCALBIN)/gosec || GOBIN=$(LOCALBIN) go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)

OPSDK = ./bin/operator-sdk
.PHONY: operator-sdk
operator-sdk: ## Download operator-sdk locally if necessary.
ifeq (,$(wildcard $(OPSDK)))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPSDK)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	echo "Dowloading https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH} to ./bin/operator-sdk" ;\
	curl -sSLo $(OPSDK) https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH} ;\
	chmod +x $(OPSDK) ;\
	}
endif

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	echo "Dowloading https://github.com/operator-framework/operator-registry/releases/download/$(OPM_VERSION)/$${OS}-$${ARCH}-opm to ./bin/opm" ;\
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/$(OPM_VERSION)/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
endif

## this will fail if VERSION is not a semver x.y.z version
# Build a bundle image using the 'operator-sdk' tool
.PHONY: bundle
bundle: operator-sdk manifests kustomize ## Generate bundle manifests and metadata, then validate generated files.
	$(OPSDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | $(OPSDK) generate bundle $(BUNDLE_GEN_FLAGS)
	$(OPSDK) bundle validate ./bundle
	mv -f bundle.Dockerfile docker/bundle.Dockerfile
	$(MAKE) fmt_license

## to update the CSV with a new tagged version of the operator:
## yq '.spec.install.spec.deployments[0].spec.template.spec.containers[1].image|="quay.io/rhdh/operator:some-other-tag"' bundle/manifests/backstage-operator.clusterserviceversion.yaml
## or 
## sed -r -e "s#(image: +)quay.io/.+operator.+#\1quay.io/rhdh/operator:some-other-tag#g" -i bundle/manifests/backstage-operator.clusterserviceversion.yaml
.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	$(CONTAINER_ENGINE) build --platform $(PLATFORM) -f docker/bundle.Dockerfile -t $(BUNDLE_IMG) --label $(LABEL) .

.PHONY: bundle-push
bundle-push: ## Push bundle image to registry
	$(MAKE) image-push IMG=$(BUNDLE_IMG)

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: bundle-push opm ## Generate operator-catalog dockerfile using the operator-bundle image built and published above; then build catalog image
    ## [GA] added '-d docker/index.Dockerfile' to avoid generating in the root
	$(OPM) index add --container-tool $(CONTAINER_ENGINE) --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT) --generate -d docker/index.Dockerfile
	$(CONTAINER_ENGINE) build --platform $(PLATFORM) -f docker/index.Dockerfile -t $(CATALOG_IMG) --label $(LABEL) .

.PHONY: catalog-push
catalog-push: ## Push catalog image to registry
	$(MAKE) image-push IMG=$(CATALOG_IMG)

.PHONY:
release-build: bundle image-build bundle-build catalog-build ## Build operator, bundle + catalog images

.PHONY:
release-push: image-push bundle-push catalog-push ## Push operator, bundle + catalog images

.PHONY: deploy-olm
deploy-olm: ## Deploy the operator with OLM
	kubectl apply -f config/samples/catalog-operator-group.yaml
	sed "s/{{VERSION}}/$(subst /,\/,$(VERSION))/g" config/samples/catalog-subscription-template.yaml | sed "s/{{DEFAULT_OLM_NAMESPACE}}/$(subst /,\/,$(DEFAULT_OLM_NAMESPACE))/g" | kubectl apply -f -

.PHONY: undeploy-olm
undeploy-olm: ## Un-deploy the operator with OLM
	-kubectl delete subscriptions.operators.coreos.com backstage-operator
	-kubectl delete operatorgroup backstage-operator-group
	-kubectl delete clusterserviceversion backstage-operator.$(VERSION)

DEFAULT_OLM_NAMESPACE ?= openshift-marketplace
.PHONY: catalog-update
catalog-update: ## Update catalog source in the default namespace for catalogsource
	-kubectl delete catalogsource backstage-operator -n $(DEFAULT_OLM_NAMESPACE)
	sed "s/{{CATALOG_IMG}}/$(subst /,\/,$(CATALOG_IMG))/g" config/samples/catalog-source-template.yaml | kubectl apply -n $(DEFAULT_OLM_NAMESPACE) -f -

.PHONY: deploy-openshift
deploy-openshift: release-build release-push catalog-update ## Deploy the operator on openshift cluster

