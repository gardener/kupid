VERSION             := $(shell cat VERSION)
REGISTRY            := europe-docker.pkg.dev/gardener-project/public

IMAGE_REPOSITORY    := $(REGISTRY)/gardener/kupid
IMAGE_TAG           := $(VERSION)
REPO_ROOT           := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_REPOSITORY):$(IMAGE_TAG)

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
CONTROLLER_GEN_REQVERSION := v0.19.0

# Set cert-manager version to deploy
CERTMANAGER_VERSION := v1.12.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

include hack/tools.mk

set-permissions:
	@chmod +x $(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/*

all: webhook

revendor: set-permissions
	@env GO111MODULE=on go mod tidy -v
	@env GO111MODULE=on go mod vendor -v
	@make set-permissions

update-dependencies:
	@env GO111MODULE=on go get -u
	@make revendor

verify: check test

# Run checks
check:
	@.ci/check

.PHONY: sast
sast: $(GOSEC)
	@./hack/sast.sh

.PHONY: sast-report
sast-report: $(GOSEC)
	@./hack/sast.sh --gosec-report true

# Run tests
test: generate fmt vet manifests
	@.ci/test

# Build webhook binary
webhook: generate fmt vet
	@.ci/build

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run -mod vendor ./main.go --kubeconfig "${KUBECONFIG}"

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/webhook && kustomize edit set image kupid=${IMG}
	kustomize build config/default | kubectl apply -f -

# Undeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy: manifests
	kustomize build config/default | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy-with-certmanager: manifests
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/${CERTMANAGER_VERSION}/cert-manager.yaml
	kubectl -n cert-manager wait --for=condition=Ready pod -l app=webhook,app.kubernetes.io/instance=cert-manager
	cd config/webhook && kustomize edit set image kupid=${IMG}
	kustomize build config/with-certmanager | kubectl apply -f -

# Undeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy-with-certmanager: manifests
	kustomize build config/with-certmanager | kubectl delete -f -
	kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/${CERTMANAGER_VERSION}/cert-manager.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: set-permissions controller-gen
	"$(CONTROLLER_GEN)" $(CRD_OPTIONS) webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: set-permissions controller-gen
	"$(CONTROLLER_GEN)" object:headerFile="hack/boilerplate.go.txt" paths="./..."
	@./vendor/github.com/gardener/gardener/hack/generate.sh ./charts/...

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_REQVERSION) ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else ifneq (Version: $(CONTROLLER_GEN_REQVERSION), $(shell controller-gen --version))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_REQVERSION) ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
