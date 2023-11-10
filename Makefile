# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.1.0

APP_NAME ?= backstage-operator

CONTAINER_ENGINE ?= docker

LOCALDIR ?= $(shell pwd)
## Location to install dependencies to
LOCALBIN ?= $(LOCALDIR)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CRD_REF_DOCS ?= $(LOCALBIN)/crd-ref-docs
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
OPM_VERSION ?= v1.26.2
OPERATOR_SDK_VERSION ?= v1.25.0
CONTROLLER_TOOLS_VERSION ?= v0.13.0
ENVTEST_K8S_VERSION ?= 1.25.0
KUSTOMIZE_VERSION ?= v4.5.7
GOIMPORTS_VERSION ?= v0.1.11
REVIVE_VERSION ?= v1.2.1
CRD_REF_DOCS_VERSION ?= v0.0.8

# Registry for docker images
REGISTRY ?= quay.io/rhdh

# CATALOG_BASE_IMG defines an existing catalog version to build on & add bundles to
CATALOG_BASE_IMG ?= $(REGISTRY)/$(APP_NAME)-catalog:v$(VERSION)

export OPERATOR_CONDITION_NAME=$(APP_NAME):v$(VERSION)

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "preview,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=preview,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="preview,fast,stable")
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
IMAGE_TAG_BASE ?= $(REGISTRY)/$(APP_NAME)

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# OLD_BUNDLE_IMGS defines the comma separated list of old bundles to add to the index.
COMMA := ,
EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
OLD_BUNDLE_IMG_TAG_BASE ?= $(IMAGE_TAG_BASE)-bundle
OLD_BUNDLE_IMGS ?= $(patsubst %$(COMMA),%$(EMPTY),$(subst $(SPACE),$(EMPTY),$(foreach ver,$(subst $(COMMA),$(SPACE),$(OLD_BUNDLE_VERSIONS)),$(OLD_BUNDLE_IMG_TAG_BASE):v$(ver),)))

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_TAG_BASE):v$(VERSION)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
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
	$(CONTROLLER_GEN) crd rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: install-tools
install-tools:
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	go install github.com/mgechev/revive@$(REVIVE_VERSION)

.PHONY: fmt
fmt: install-tools
	goimports -w .

.PHONY: lint
lint: install-tools
	goimports -d .
	revive -config ./config.toml ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: test
test: sdk-manifests vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.tmp.out -covermode count

.PHONY: coverage
coverage: test
	cat cover.tmp.out | grep -v "_generated.*.go" > cover.out
	go tool cover -func=cover.out

##@ Build
.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: sdk-manifests vet ## Run a controller from your host.
	go run ./cmd/main.go

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_ENGINE) build --pull --platform linux/amd64 -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_ENGINE) push ${IMG}

.PHONY: release-build
release-build: bundle docker-build bundle-build catalog-build ## Build operator docker, bundle, catalog images

.PHONY: release-push
release-push: docker-push bundle-push catalog-push ## Push operator docker, bundle, catalog images

##@ Deployment
.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

.PHONY: deploy-olm
deploy-olm: ## Deploy the operator with OLM
	oc apply -f config/samples/catalog-operator-group.yaml
	oc apply -f config/samples/catalog-subscription.yaml

.PHONY: undeploy-olm
undeploy-olm: ## Un-deploy the operator with OLM
	-oc delete subscriptions.operators.coreos.com $(APP_NAME)
	-oc delete operatorgroup $(APP_NAME)-group
	-oc delete clusterserviceversion backstage-deploy-operator.v${VERSION}

.PHONY: catalog-update
catalog-update: ## Update catalog source in namespace openshift-marketplace 
	-oc delete catalogsource $(APP_NAME) -n openshift-marketplace
	sed "s/{{CATALOG_IMG}}/$(subst /,\/,$(CATALOG_IMG))/g" config/samples/catalog-source-template.yaml | oc apply -f -

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: crd-ref-docs
crd-ref-docs: $(CRD_REF_DOCS) ## Download crd-ref-docs locally if necessary.
$(CRD_REF_DOCS): $(LOCALBIN)
	test -s $(LOCALBIN)/crd-ref-docs || GOBIN=$(LOCALBIN) go install github.com/elastic/crd-ref-docs@$(CRD_REF_DOCS_VERSION)

.PHONY: generate-ref
generate-ref: generate fmt crd-ref-docs
	$(CRD_REF_DOCS) --log-level=WARN --config=$(LOCALDIR)/config/ref-templates/config.yaml --source-path=$(LOCALDIR)/api/v1alpha1 --renderer=markdown --templates-dir=$(LOCALDIR)/ref-templates/markdown --output-path=$(LOCALDIR)/docs/api/markdown/ref.md

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v4@$(KUSTOMIZE_VERSION)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: sdk-manifests
sdk-manifests: manifests kustomize sdk ## Generate bundle manifests and metadata.
	$(SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)

.PHONY: bundle
bundle: sdk-manifests ## Generate bundle manifests, then validate generated files.
	$(KUSTOMIZE) build config/manifests | $(SDK) generate bundle -q --overwrite --manifests --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(SDK) bundle validate ./bundle --select-optional suite=operatorframework

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	$(CONTAINER_ENGINE) build -f bundle.Dockerfile --platform linux/amd64 -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = ./bin/opm.$(OPM_VERSION)
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/$(OPM_VERSION)/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
endif

.PHONY: sdk
SDK = ./bin/operator-sdk.$(OPERATOR_SDK_VERSION)
sdk: ## Download operator-sdk if necessary.
ifeq (,$(wildcard $(SDK)))
	@{ \
	set -e ;\
	mkdir -p $(dir $(SDK)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(SDK) https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH} ;\
	chmod +x $(SDK) ;\
	}
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
ifeq ($(OLD_BUNDLE_IMGS),)
BUNDLE_IMGS ?= $(BUNDLE_IMG)
else
BUNDLE_IMGS ?= $(BUNDLE_IMG),$(OLD_BUNDLE_IMGS)
endif

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm bundle-push ## Build a catalog image. The bundle image must have been pushed into the registry.git st
	$(OPM) index add --container-tool $(CONTAINER_ENGINE) --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)

.PHONY: get-version
get-version: ; $(info ${VERSION})
	@echo -n
