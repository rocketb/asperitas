# Adapted from https://www.thapaliya.com/en/writings/well-documented-makefiles/
.PHONY: help
help: ## Display this help and any documented user-facing targets. Other undocumented targets may be present in the Makefile.
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-45s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := all
.PHONY: all asperitas-api metrics asperitas-admin

SHELL = /usr/bin/env bash -o pipefail

GOTEST ?= go test

################
# Main Targets #
################
all: asperitas-api metrics asperitas-admin ## build all executables

#################
# asperitas api #
#################
.PHONY: cmd/asperitas/api/asperitas-api
asperitas-api: cmd/asperitas/api/asperitas-api ## build asperitas API

cmd/asperitas/api/asperitas-api:
	CGO_ENABLED=0 go build -mod vendor -o $@ ./$(@D)

###########
# metrics #
###########
.PHONY: cmd/asperitas/metrics/metrics
metrics: cmd/asperitas/metrics/metrics ## build app metrics processing endpoints

cmd/asperitas/metrics/metrics:
	CGO_ENABLED=0 go build -mod vendor -o $@ ./$(@D)

###############
# admin-tools #
###############
.PHONY: cmd/tools/asperitas-admin/asperitas-admin
asperitas-admin: cmd/tools/asperitas-admin/asperitas-admin ## build admin cmd tool

cmd/tools/asperitas-admin/asperitas-admin:
	CGO_ENABLED=0 go build -mod vendor -o $@ ./$(@D)

########
# Lint #
########
lint: ## run linters
	go version
	golangci-lint version
	GO111MODULE=on golangci-lint run -v
	# faillint -paths "sync/atomic=go.uber.org/atomic" ./...

########
# Test #
########

test: ## run the unit tests
	$(GOTEST) -covermode=atomic -coverprofile=cover.out -p=4 ./... | sed "s:$$: ${DRONE_STEP_NAME} ${DRONE_SOURCE_BRANCH}:" | tee test_results.txt

#########
# Clean #
#########
clean: ## clean the generated files
	rm -rf cmd/asperitas/api/asperitas-api
	rm -rf cmd/asperitas/metrics/metrics
	rm -rf cmd/tools/asperitas-admin/asperitas-admin
	go clean ./...

########
# Misc #
########

# support go modules
check-mod:
	@git diff --exit-code -- go.sum go.mod vendor/ || \
	    (echo "Run 'go mod download && go mod verify && go mod tidy && go mod vendor' and check in changes to vendor/ to fix failed check-mod."; exit 1)

format:
	find . -type f -name '*.go' -exec gofmt -w -s {} \;
	find . -type f -name '*.go' -exec goimports -w -local github.com/rocketb/asperitas {} \;

tidy:
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache
