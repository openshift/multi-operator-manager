all: build
.PHONY: all

# Include the library makefile
include $(addprefix ./vendor/github.com/openshift/build-machinery-go/make/, \
	golang.mk \
	targets/openshift/deps-gomod.mk \
	targets/openshift/images.mk \
	targets/openshift/bindata.mk \
)

# Run core verification and all self contained tests.
#
# Example:
#   make check
check: | verify test-unit
.PHONY: check

IMAGE_REGISTRY?=registry.ci.openshift.org

# This will call a macro called "build-image" which will generate image specific targets based on the parameters:
# $0 - macro name
# $1 - target name
# $2 - image ref
# $3 - Dockerfile path
# $4 - context directory for image build
# It will generate target "image-$(1)" for building the image and binding it as a prerequisite to target "images".
$(call build-image,multi-operator-manager,$(IMAGE_REGISTRY)/ocp/4.18:multi-operator-manager,./Dockerfile.openshift,.)

clean:
	$(RM) multi-operator-manager
	$(RM) -rf test-output
.PHONY: clean

GO_TEST_PACKAGES :=./pkg/... ./cmd/...

# Run e2e tests. Requires openshift-tests in $PATH.
#
# Example:
#   make test-e2e
test-e2e:
	hack/e2e.sh

.PHONY: test-e2e

test-operator-integration: build
	hack/test-operator-integration.sh

.PHONY: test-operator-integration

update-test-operator-integration: build
	REPLACE_TEST_OUTPUT=true hack/test-operator-integration.sh

.PHONY: update-test-operator-integration
