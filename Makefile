# provide JUSTFLAGS for just-backed targets
include ./justfiles/flags.mk

BEDROCK_TAGS_REMOTE?=origin
OP_STACK_GO_BUILDER?=us-docker.pkg.dev/oplabs-tools-artifacts/images/op-stack-go:latest

# Requires at least Python v3.9; specify a minor version below if needed
PYTHON?=python3

help: ## Prints this help message
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: build-go build-contracts ## Builds Go components and contracts-bedrock
.PHONY: build

build-go: submodules cp-node cp-deployer## Builds cp-node and cp-deployer
.PHONY: build-go

build-contracts:
	(cd packages/contracts-bedrock && just build)
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
			cp-node op-batcher op-proposer op-challenger op-dispute-mon cp-supervisor
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

contracts-bedrock-docker: ## Builds Docker image for Bedrock contracts
	IMAGE_TAGS=$$(git rev-parse HEAD),latest \
	docker buildx bake \
			--progress plain \
			--load \
			-f docker-bake.hcl \
		  contracts-bedrock
.PHONY: contracts-bedrock-docker

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

op-batcher: ## Builds op-batcher binary
	just $(JUSTFLAGS) ./op-batcher/op-batcher
.PHONY: op-batcher

op-proposer: ## Builds op-proposer binary
	just $(JUSTFLAGS) ./op-proposer/op-proposer
.PHONY: op-proposer

cp-deployer: ## Builds cp-deployer binary
	just $(JUSTFLAGS) ./cp-deployer/build
.PHONY: cp-deployer

op-challenger: ## Builds op-challenger binary
	make -C ./op-challenger op-challenger
.PHONY: op-challenger

op-dispute-mon: ## Builds op-dispute-mon binary
	make -C ./op-dispute-mon op-dispute-mon
.PHONY: op-dispute-mon

cp-program: ## Builds cp-program binary
	make -C ./cp-program cp-program
.PHONY: cp-program

cannon:  ## Builds cannon binary
	make -C ./cannon cannon
.PHONY: cannon

reproducible-prestate:   ## Builds reproducible-prestate binary
	make -C ./cp-program reproducible-prestate
.PHONY: reproducible-prestate

# Include any files required for the devnet to build and run.
DEVNET_CANNON_PRESTATE_FILES := cp-program/bin/prestate-proof.json cp-program/bin/prestate.bin.gz cp-program/bin/prestate-proof-mt64.json cp-program/bin/prestate-mt64.bin.gz cp-program/bin/prestate-interop.bin.gz


$(DEVNET_CANNON_PRESTATE_FILES):
	make cannon-prestate
	make cannon-prestate-mt64
	make cannon-prestate-interop

cannon-prestates: cannon-prestate cannon-prestate-mt64 cannon-prestate-interop
.PHONY: cannon-prestates

cannon-prestate: cp-program cannon ## Generates prestate using cannon and cp-program
	./cannon/bin/cannon load-elf --type singlethreaded-2 --path cp-program/bin/cp-program-client.elf --out cp-program/bin/prestate.bin.gz --meta cp-program/bin/meta.json
	./cannon/bin/cannon run --proof-at '=0'  --stop-at '=1' --input cp-program/bin/prestate.bin.gz --meta cp-program/bin/meta.json --proof-fmt 'cp-program/bin/%d.json' --output ""
	mv cp-program/bin/0.json cp-program/bin/prestate-proof.json
.PHONY: cannon-prestate

cannon-prestate-mt64: cp-program cannon ## Generates prestate using cannon and cp-program in the latest 64-bit multithreaded cannon format
	./cannon/bin/cannon load-elf --type multithreaded64-3 --path cp-program/bin/cp-program-client64.elf --out cp-program/bin/prestate-mt64.bin.gz --meta cp-program/bin/meta-mt64.json
	./cannon/bin/cannon run --proof-at '=0' --stop-at '=1' --input cp-program/bin/prestate-mt64.bin.gz --meta cp-program/bin/meta-mt64.json --proof-fmt 'cp-program/bin/%d-mt64.json' --output ""
	mv cp-program/bin/0-mt64.json cp-program/bin/prestate-proof-mt64.json
.PHONY: cannon-prestate-mt64

cannon-prestate-interop: cp-program cannon ## Generates interop prestate using cannon and cp-program in the latest 64-bit multithreaded cannon format
	./cannon/bin/cannon load-elf --type multithreaded64-3 --path cp-program/bin/cp-program-client-interop.elf --out cp-program/bin/prestate-interop.bin.gz --meta cp-program/bin/meta-interop.json
	./cannon/bin/cannon run --proof-at '=0' --stop-at '=1' --input cp-program/bin/prestate-interop.bin.gz --meta cp-program/bin/meta-interop.json --proof-fmt 'cp-program/bin/%d-interop.json' --output ""
	mv cp-program/bin/0-interop.json cp-program/bin/prestate-proof-interop.json
.PHONY: cannon-prestate-interop

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
	cd packages/contracts-bedrock/ && forge clean
.PHONY: clean

nuke: clean ## Completely clean the project directory
	git clean -Xdf
.PHONY: nuke

test-unit: ## Runs unit tests for all components
	make -C ./cp-node test
	make -C ./op-proposer test
	make -C ./op-batcher test
	make -C ./op-e2e test
	(cd packages/contracts-bedrock && just test)
.PHONY: test-unit

# Remove the baseline-commit to generate a base reading & show all issues
semgrep: ## Runs Semgrep checks
	$(eval DEV_REF := $(shell git rev-parse develop))
	SEMGREP_REPO_NAME=/cpchain-network/cp-chain semgrep ci --baseline-commit=$(DEV_REF)
.PHONY: semgrep

update-op-geth: ## Updates the Geth version used in the project
	./ops/scripts/update-op-geth.py
.PHONY: update-op-geth
