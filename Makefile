# provide JUSTFLAGS for just-backed targets
include ./justfiles/flags.mk

# Requires at least Python v3.9; specify a minor version below if needed
PYTHON?=python3

help: ## Prints this help message
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: build-go build-contracts ## Builds Go components and contracts-cpchain
.PHONY: build

build-go: submodules cp-node cp-deployer## Builds cp-node and cp-deployer
.PHONY: build-go

build-contracts:
	(cd packages/contracts-cpchain && just build)
.PHONY: build-contracts

lint-go: ## Lints Go code with specific linters
	golangci-lint run -E goimports,sqlclosecheck,bodyclose,asciicheck,misspell,errorlint --timeout 5m -e "errors.As" -e "errors.Is" ./...
	golangci-lint run -E err113 --timeout 5m -e "errors.As" -e "errors.Is" ./cp-program/client/...
.PHONY: lint-go

lint-go-fix: ## Lints Go code with specific linters and fixes reported issues
	golangci-lint run -E goimports,sqlclosecheck,bodyclose,asciicheck,misspell,errorlint --timeout 5m -e "errors.As" -e "errors.Is" ./... --fix
.PHONY: lint-go-fix

golang-docker: ## Builds Docker images for Go components using buildx
	# We don't use a buildx builder here, and just load directly into regular docker, for convenience.
	GIT_COMMIT=$$(git rev-parse HEAD) \
	GIT_DATE=$$(git show -s --format='%ct') \
	IMAGE_TAGS=$$(git rev-parse HEAD),latest \
	docker buildx bake \
			--progress plain \
			--load \
			-f docker-bake.hcl \
			cp-node cp-supervisor
.PHONY: golang-docker

docker-builder-clean: ## Removes the Docker buildx builder
	docker buildx rm buildx-build
.PHONY: docker-builder-clean

docker-builder: ## Creates a Docker buildx builder
	docker buildx create \
		--driver=docker-container --name=buildx-build --bootstrap --use
.PHONY: docker-builder

# add --print to dry-run
cross-cp-node: ## Builds cross-platform Docker image for cp-node
	# We don't use a buildx builder here, and just load directly into regular docker, for convenience.
	GIT_COMMIT=$$(git rev-parse HEAD) \
	GIT_DATE=$$(git show -s --format='%ct') \
	IMAGE_TAGS=$$(git rev-parse HEAD),latest \
	PLATFORMS="linux/arm64" \
	GIT_VERSION=$(shell tags=$$(git tag --points-at $(GITCOMMIT) | grep '^cp-node/' | sed 's/cp-node\///' | sort -V); \
             preferred_tag=$$(echo "$$tags" | grep -v -- '-rc' | tail -n 1); \
             if [ -z "$$preferred_tag" ]; then \
                 if [ -z "$$tags" ]; then \
                     echo "untagged"; \
                 else \
                     echo "$$tags" | tail -n 1; \
                 fi \
             else \
                 echo $$preferred_tag; \
             fi) \
	docker buildx bake \
			--progress plain \
			--builder=buildx-build \
			--load \
			--no-cache \
			-f docker-bake.hcl \
			cp-node
.PHONY: cross-cp-node

contracts-cpchain-docker: ## Builds Docker image for Bedrock contracts
	IMAGE_TAGS=$$(git rev-parse HEAD),latest \
	docker buildx bake \
			--progress plain \
			--load \
			-f docker-bake.hcl \
		  contracts-cpchain
.PHONY: contracts-cpchain-docker

submodules: ## Updates git submodules
	git submodule update --init --recursive
.PHONY: submodules


cp-node: ## Builds cp-node binary
	just $(JUSTFLAGS) ./cp-node/cp-node
.PHONY: cp-node

generate-mocks-cp-node: ## Generates mocks for cp-node
	make -C ./cp-node generate-mocks
.PHONY: generate-mocks-cp-node

generate-mocks-cp-service: ## Generates mocks for cp-service
	make -C ./cp-service generate-mocks
.PHONY: generate-mocks-cp-service

cp-deployer: ## Builds cp-deployer binary
	just $(JUSTFLAGS) ./cp-deployer/build
.PHONY: cp-deployer

cp-program: ## Builds cp-program binary
	make -C ./cp-program cp-program
.PHONY: cp-program

reproducible-prestate:   ## Builds reproducible-prestate binary
	make -C ./cp-program reproducible-prestate
.PHONY: reproducible-prestate

mod-tidy: ## Cleans up unused dependencies in Go modules
	# Below GOPRIVATE line allows mod-tidy to be run immediately after
	# releasing new versions. This bypasses the Go modules proxy, which
	# can take a while to index new versions.
	#
	# See https://proxy.golang.org/ for more info.
	export GOPRIVATE="github.com/ethereum-optimism" && go mod tidy
.PHONY: mod-tidy

clean: ## Removes all generated files under bin/
	rm -rf ./bin
	cd packages/contracts-cpchain/ && forge clean
.PHONY: clean

nuke: clean ## Completely clean the project directory
	git clean -Xdf
.PHONY: nuke

test-unit: ## Runs unit tests for all components
	make -C ./cp-node test
	(cd packages/contracts-cpchain && just test)
.PHONY: test-unit

# Remove the baseline-commit to generate a base reading & show all issues
semgrep: ## Runs Semgrep checks
	$(eval DEV_REF := $(shell git rev-parse develop))
	SEMGREP_REPO_NAME=/cpchain-network/cp-chain semgrep ci --baseline-commit=$(DEV_REF)
.PHONY: semgrep

update-cp-geth: ## Updates the Geth version used in the project
	./ops/scripts/update-cp-geth.py
.PHONY: update-cp-geth
