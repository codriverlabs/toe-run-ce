# Image URLs
IMG ?= controller:latest
CONTROLLER_IMG ?= ghcr.io/codriverlabs/toe-controller:$(VERSION)
COLLECTOR_IMG ?= ghcr.io/codriverlabs/toe-collector:$(VERSION)
APERF_IMG ?= ghcr.io/codriverlabs/toe-aperf:$(VERSION)
VERSION ?= v1.1.5-beta

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
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
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet setup-envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test $$(go list ./... | grep -v /e2e) -coverprofile cover.out

# TODO(user): To use a different vendor for e2e tests, modify the setup under 'tests/e2e'.
# The default setup assumes Kind is pre-installed and builds/loads the Manager Docker image locally.
# CertManager is installed by default; skip with:
# - CERT_MANAGER_INSTALL_SKIP=true
KIND_CLUSTER ?= toe-test-e2e

.PHONY: setup-test-e2e
setup-test-e2e: ## Set up a Kind cluster for e2e tests if it does not exist
	@command -v $(KIND) >/dev/null 2>&1 || { \
		echo "Kind is not installed. Please install Kind manually."; \
		exit 1; \
	}
	@case "$$($(KIND) get clusters)" in \
		*"$(KIND_CLUSTER)"*) \
			echo "Kind cluster '$(KIND_CLUSTER)' already exists. Skipping creation." ;; \
		*) \
			echo "Creating Kind cluster '$(KIND_CLUSTER)'..."; \
			$(KIND) create cluster --name $(KIND_CLUSTER) ;; \
	esac

.PHONY: test-e2e
test-e2e: setup-test-e2e manifests generate fmt vet ## Run the e2e tests. Expected an isolated environment using Kind.
	KIND=$(KIND) KIND_CLUSTER=$(KIND_CLUSTER) go test -tags=e2e ./test/e2e/ -v -ginkgo.v
	$(MAKE) cleanup-test-e2e

.PHONY: cleanup-test-e2e
cleanup-test-e2e: ## Tear down the Kind cluster used for e2e tests
	@$(KIND) delete cluster --name $(KIND_CLUSTER)

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Verify golangci-lint linter configuration
	$(GOLANGCI_LINT) config verify

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

# If you wish to build the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

.PHONY: docker-build-aperf
docker-build-aperf: ## Build aperf power tool docker image.
	$(CONTAINER_TOOL) build -t ghcr.io/codriverlabs/toe-aperf:$(VERSION) power-tools/aperf/

.PHONY: docker-push-aperf
docker-push-aperf: ## Push aperf power tool docker image.
	$(CONTAINER_TOOL) push ghcr.io/codriverlabs/toe-aperf:$(VERSION)

.PHONY: docker-build-all
docker-build-all: docker-build docker-build-aperf ## Build all docker images.

.PHONY: docker-push-all
docker-push-all: docker-push docker-push-aperf ## Push all docker images.

# PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - be able to use docker buildx. More info: https://docs.docker.com/build/buildx/
# - have enabled BuildKit. More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image to your registry (i.e. if you do not set a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To adequately provide solutions that are compatible with multiple platforms, you should consider using this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name toe-builder
	$(CONTAINER_TOOL) buildx use toe-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm toe-builder
	rm Dockerfile.cross

##@ GitHub Release

.PHONY: github-release
github-release: helm-chart helm-package build-installer ## Generate all release artifacts for GitHub
	@echo "🚀 Generating GitHub release artifacts..."
	
	# Create release directory
	mkdir -p dist/release
	
	# Copy installer YAML
	cp dist/install.yaml dist/release/toe-operator-$(VERSION).yaml
	
	# Copy Helm package
	cp dist/helm/toe-operator-*.tgz dist/release/
	
	# Generate checksums
	cd dist/release && sha256sum * > checksums.txt
	
	@echo "✅ GitHub release artifacts ready in dist/release/"
	@echo ""
	@echo "📦 Release files:"
	@ls -la dist/release/
	@echo ""
	@echo "🔗 Usage:"
	@echo "  # Direct YAML install:"
	@echo "  kubectl apply -f https://github.com/codriverlabs/toe/releases/download/v$(VERSION)/toe-operator-$(VERSION).yaml"
	@echo ""
	@echo "  # Helm install:"
	@echo "  helm install toe-operator https://github.com/codriverlabs/toe/releases/download/v$(VERSION)/toe-operator-$(VERSION).tgz"

.PHONY: github-release-controller
github-release-controller: ## Generate controller-only release artifacts
	@echo "🎯 Generating controller-only release..."
	$(MAKE) github-release IMG=ghcr.io/codriverlabs/toe-controller:$(VERSION)

.PHONY: github-release-collector  
github-release-collector: ## Generate collector-only release artifacts
	@echo "📊 Generating collector-only release..."
	mkdir -p dist/release
	$(KUSTOMIZE) build deploy/collector > dist/release/toe-collector-$(VERSION).yaml
	@echo "✅ Collector release ready: dist/release/toe-collector-$(VERSION).yaml"

##@ Release

.PHONY: helm-chart
helm-chart: manifests generate kustomize ## Generate Helm chart with configurable values
	@echo "🔨 Generating Helm chart..."
	
	# Copy Helm chart structure
	mkdir -p dist/helm
	cp -r helm/toe-operator dist/helm/
	
	# Update image tags in values.yaml based on build parameters
	@CONTROLLER_TAG=$$(echo "$(CONTROLLER_IMG)" | sed 's/.*://'); \
	COLLECTOR_TAG=$$(echo "$(COLLECTOR_IMG)" | sed 's/.*://'); \
	APERF_TAG=$$(echo "$(APERF_IMG)" | sed 's/.*://'); \
	sed -i "s|tag: \"1.1.4-beta\"|tag: \"$$CONTROLLER_TAG\"|g" dist/helm/toe-operator/values.yaml
	
	# Generate CRDs for Helm (installed before templates)
	mkdir -p dist/helm/toe-operator/crds
	$(KUSTOMIZE) build config/crd > dist/helm/toe-operator/crds/crds.yaml
	
	# Generate controller manifests with Helm templating (excluding CRDs and Namespace)
	mkdir -p dist/helm/toe-operator/templates
	cd config/manager && $(KUSTOMIZE) edit set image controller='{{ include "toe-operator.controller.image" . }}'
	$(KUSTOMIZE) build config/default > /tmp/kustomize-output.yaml
	sed 's/namespace: toe-system/namespace: {{ include "toe-operator.namespace" . }}/g' /tmp/kustomize-output.yaml | \
		sed '/^apiVersion: apiextensions\.k8s\.io\/v1$$/,/^---$$/d' | \
		sed "s|'{{ include \"toe-operator.controller.image\" . }}'|{{ include \"toe-operator.controller.image\" . }}|g" | \
		sed "s|{{ include \"toe-operator.controller.image\" . }}:latest|{{ include \"toe-operator.controller.image\" . }}|g" > dist/helm/toe-operator/templates/controller.yaml
	rm -f /tmp/kustomize-output.yaml
	
	# Add Helm conditionals to controller template
	sed -i '1i{{- if .Values.controller.enabled }}' dist/helm/toe-operator/templates/controller.yaml
	echo '{{- end }}' >> dist/helm/toe-operator/templates/controller.yaml
	
	# Add imagePullSecrets to controller deployment if needed
	sed -i '/serviceAccountName:/a\      {{- with include "toe-operator.imagePullSecrets" . }}\n      imagePullSecrets:\n{{ . | indent 8 }}\n      {{- end }}' dist/helm/toe-operator/templates/controller.yaml
	
	# Remove any hardcoded imagePullSecrets from kustomize output
	sed -i '/^      imagePullSecrets:$$/,/^      securityContext:$$/{/^      imagePullSecrets:$$/d; /^      - name: /d;}' dist/helm/toe-operator/templates/controller.yaml
	
	# Reset kustomize to original image
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	
	@echo "✅ Helm chart generated in dist/helm/toe-operator/"

.PHONY: helm-package
helm-package: helm-chart ## Package Helm chart into .tgz file
	@command -v helm >/dev/null 2>&1 || { echo "❌ Helm is required. Install from https://helm.sh/docs/intro/install/"; exit 1; }
	# Update Chart.yaml version before packaging
	sed -i 's/^version: .*/version: $(VERSION)/' dist/helm/toe-operator/Chart.yaml
	sed -i 's/^appVersion: .*/appVersion: "$(VERSION)"/' dist/helm/toe-operator/Chart.yaml
	# Copy ECR sync script to helm chart directory
	mkdir -p dist/helm/toe-operator/scripts
	cp helper_scripts/ecr/sync-images-from-ghcr-to-ecr.sh dist/helm/toe-operator/scripts/
	# Copy examples folder and update image references
	mkdir -p dist/helm/toe-operator/examples
	cp -r examples/* dist/helm/toe-operator/examples/
	# Update aperf image reference in powertoolconfig-aperf.yaml
	if [ -f "dist/helm/toe-operator/examples/powertoolconfig-aperf.yaml" ]; then \
		sed -i 's|image: .*|image: $(APERF_IMG)|g' dist/helm/toe-operator/examples/powertoolconfig-aperf.yaml; \
	fi
	cd dist/helm && helm package toe-operator
	@echo "✅ Helm chart packaged: dist/helm/toe-operator-$(VERSION).tgz"

.PHONY: render-locally-helm-chart
render-locally-helm-chart: ## Render Helm chart locally with custom values for testing
	@echo "🎨 Rendering Helm chart locally..."
	
	# Check required parameters
	@if [ -z "$(AWS_ACCOUNT_ID)" ]; then \
		echo "❌ Error: AWS_ACCOUNT_ID is required. Usage: make render-locally-helm-chart AWS_ACCOUNT_ID=123456789012 AWS_REGION=us-west-2"; \
		exit 1; \
	fi
	@if [ -z "$(AWS_REGION)" ]; then \
		echo "❌ Error: AWS_REGION is required. Usage: make render-locally-helm-chart AWS_ACCOUNT_ID=123456789012 AWS_REGION=us-west-2"; \
		exit 1; \
	fi
	
	# Clean up previous artifacts
	rm -rf ./tmp/*
	rm -rf ./dist/*
	
	# Generate and package Helm chart
	$(MAKE) manifests
	$(MAKE) helm-chart
	$(MAKE) helm-package
	
	# Extract and render chart with custom values
	mkdir -p ./tmp
	cp ./dist/helm/*.tgz ./tmp
	cd ./tmp && tar -xf ./*.tgz
	cd ./tmp && helm template \
		--set-string global.registry.repository=$(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/codriverlabs/toe \
		--set-string controller.image.tag=$(VERSION) \
		--set-string collector.image.tag=$(VERSION) \
		toe-operator-$(VERSION) ./toe-operator > template.yaml
	
	@echo "✅ Helm chart rendered locally: tmp/template.yaml"
	@echo "📦 Extracted chart available in: tmp/toe-operator/"
	@echo "🔧 Used registry: $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/codriverlabs/toe"

.PHONY: build-installer
build-installer: manifests generate kustomize ## Generate a consolidated YAML with CRDs and deployment.
	mkdir -p dist
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > dist/install.yaml

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy-controller-only
undeploy-controller-only: kustomize ## Undeploy only controller resources, preserving namespace and other components
	$(KUSTOMIZE) build config/default | grep -v "kind: Namespace" | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f - --cascade=orphan || true
	# Remove namespace separately only if it's empty (preserves collector)
	@echo "Checking if namespace can be safely removed..."
	@if kubectl get all -n toe-system 2>/dev/null | grep -q "No resources found"; then \
		echo "Namespace is empty, removing..."; \
		kubectl delete namespace toe-system --ignore-not-found=true; \
	else \
		echo "Namespace contains other resources (like collector), preserving namespace"; \
		kubectl get all -n toe-system; \
	fi

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KIND ?= kind
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
KUSTOMIZE_VERSION ?= v5.6.0
CONTROLLER_TOOLS_VERSION ?= v0.18.0
#ENVTEST_VERSION is the version of controller-runtime release branch to fetch the envtest setup script (i.e. release-0.20)
ENVTEST_VERSION ?= $(shell go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime | awk -F'[v.]' '{printf "release-%d.%d", $$2, $$3}')
#ENVTEST_K8S_VERSION is the version of Kubernetes to use for setting up ENVTEST binaries (i.e. 1.31)
ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{printf "1.%d", $$3}')
GOLANGCI_LINT_VERSION ?= v2.3.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: setup-envtest
setup-envtest: envtest ## Download the binaries required for ENVTEST in the local bin directory.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@$(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $$(realpath $(1)-$(3)) $(1)
endef

# Collector targets
.PHONY: collector-build collector-push collector-deploy collector-undeploy

collector-build: ## Build the collector image
	$(CONTAINER_TOOL) build -t localhost:32000/codriverlabs/toe-collector:$(VERSION) -f build/collector/Dockerfile .

collector-push: ## Push the collector image
	$(CONTAINER_TOOL) push localhost:32000/codriverlabs/toe-collector:$(VERSION)

collector-deploy: ## Deploy the collector
	cd deploy/collector && $(KUSTOMIZE) edit set image localhost:32000/codriverlabs/toe-collector:$(VERSION)
	cd deploy/collector && $(KUSTOMIZE) build . | $(KUBECTL) apply -f -

collector-undeploy: ## Undeploy the collector
	cd deploy/collector && $(KUSTOMIZE) build . | $(KUBECTL) delete --ignore-not-found=true -f -
