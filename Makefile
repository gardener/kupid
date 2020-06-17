VERSION             := $(shell cat VERSION)
REGISTRY            := eu.gcr.io/gardener-project/gardener

IMAGE_REPOSITORY    := $(REGISTRY)/kupid
IMAGE_TAG           := $(VERSION)

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_REPOSITORY):$(IMAGE_TAG)

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: webhook

revendor:
	@env GO111MODULE=on go mod vendor -v
	@env GO111MODULE=on go mod tidy -v

update-dependencies:
	@env GO111MODULE=on go get -u
	@make revendor

verify: check test

# Run checks
check:
	@.ci/check

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
	cd config/webhook && kustomize edit set image kupid=${IMG}
	kustomize build config/with-certmanager | kubectl apply -f -

# Undeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy-with-certmanager: manifests
	kustomize build config/with-certmanager | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	"$(CONTROLLER_GEN)" $(CRD_OPTIONS) webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
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
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
