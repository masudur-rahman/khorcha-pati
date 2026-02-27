# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# The binaries to build (just the basenames).
APP_NAME = Expense Tracker
BIN := expense-tracker-bot

# Where to push the docker image.
REGISTRY ?= masudjuly02

# This version-strategy uses git tags to set the version string
git_branch       := $(shell git rev-parse --abbrev-ref HEAD)
git_tag          := $(shell git describe --exact-match --abbrev=0 2>/dev/null || echo "")
commit_hash      := $(shell git rev-parse --verify HEAD)
commit_timestamp := $(shell git show -s --format=%cd --date=format:'%Y-%m-%dT%H:%M:%S' HEAD)

VERSION          := $(shell git describe --tags --always --dirty)
version_strategy := commit_hash
ifdef git_tag
	VERSION := $(git_tag)
	version_strategy := tag
endif
#else
#	ifeq (,$(findstring $(git_branch),master HEAD))
#		ifneq (,$(patsubst release-%,,$(git_branch)))
#			VERSION := $(git_branch)
#			version_strategy := branch
#		endif
#	endif
#endif

###
### These variables should not need tweaking.
###

SRC_DIRS := api cmd infra models pkg repos services # directories which hold app source (not vendored)
EXCLUDE_DIRS := vendor,hack,bin,.go

#ALL_PLATFORMS := darwin/arm64 linux/amd64 linux/arm linux/arm64 linux/ppc64le linux/s390x windows/amd64
ALL_PLATFORMS := linux/arm64 linux/amd64

# Used internally.  Users should pass GOOS and/or GOARCH.
OS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))

#BASEIMAGE ?= gcr.io/distroless/static
BASEIMAGE ?= debian:bookworm

TAG := $(VERSION)_$(OS)_$(ARCH)

DOCKER_IMAGE := $(REGISTRY)/$(BIN)

GO_VERSION       ?= 1.24
BUILD_IMAGE      ?= ghcr.io/masudur-rahman/golang:$(GO_VERSION)

BIN_EXTENSION :=
ifeq ($(OS), windows)
  BIN_EXTENSION := .exe
endif


REPO_PKGs := user shelter pet pet_adoption
DB_TYPEs  := nosql sql

define \n


endef

# If you want to build all binaries, see the 'all-build' rule.
# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
all: # @HELP builds binaries for one platform ($OS/$ARCH)
all: build

# For the following OS/ARCH expansions, we transform OS/ARCH into OS_ARCH
# because make pattern rules don't match with embedded '/' characters.

build-%:
	@$(MAKE) build                        \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

container-%:
	@$(MAKE) container                    \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

push-%:
	@$(MAKE) push                         \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

all-build: # @HELP builds binaries for all platforms
all-build: $(addprefix build-, $(subst /,_, $(ALL_PLATFORMS)))

all-container: # @HELP builds containers for all platforms
all-container: $(addprefix container-, $(subst /,_, $(ALL_PLATFORMS)))

all-push: # @HELP pushes containers for all platforms to the defined registry
all-push: $(addprefix push-, $(subst /,_, $(ALL_PLATFORMS)))

# The following structure defeats Go's (intentional) behavior to always touch
# result files, even if they have not changed.  This will still run `go` but
# will not trigger further work if nothing has actually changed.
OUTBINS = $(foreach bin,$(BIN),bin/$(OS)_$(ARCH)/$(bin)$(BIN_EXTENSION))

build: $(OUTBINS)

# Directories that we need created to build/test.
BUILD_DIRS := bin/$(OS)_$(ARCH)     \
              .go/bin/$(OS)_$(ARCH) \
              .go/cache

# Each outbin target is just a facade for the respective stampfile target.
# This `eval` establishes the dependencies for each.
$(foreach outbin,$(OUTBINS),$(eval  \
    $(outbin): .go/$(outbin).stamp  \
))
# This is the target definition for all outbins.
$(OUTBINS):
	@true

# Each stampfile target can reference an $(OUTBIN) variable.
$(foreach outbin,$(OUTBINS),$(eval $(strip   \
    .go/$(outbin).stamp: OUTBIN = $(outbin)  \
)))
# This is the target definition for all stampfiles.
# This will build the binary under ./.go and update the real binary iff needed.
STAMPS = $(foreach outbin,$(OUTBINS),.go/$(outbin).stamp)
.PHONY: $(STAMPS)
$(STAMPS): go-build
	@echo "binary: $(OUTBIN)"
	@if ! cmp -s .go/$(OUTBIN) $(OUTBIN); then  \
	    mv .go/$(OUTBIN) $(OUTBIN);             \
	    date >$@;                               \
	fi

# This runs the actual `go build` which updates all binaries.
go-build: | $(BUILD_DIRS)
	@echo "# building for $(OS)/$(ARCH)"
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env DEBUG="$(DBG)"                                    \
	    --env GOFLAGS="$(GOFLAGS)"                              \
	    --env HTTP_PROXY="$(HTTP_PROXY)"                        \
	    --env HTTPS_PROXY="$(HTTPS_PROXY)"                      \
	    $(BUILD_IMAGE)                                          \
		/bin/bash -c "                                          \
	        ARCH=$(ARCH)                                        \
	        OS=$(OS)                                            \
	        VERSION=$(VERSION)                                  \
	        version_strategy=$(version_strategy)                \
	        git_branch=$(git_branch)                            \
	        git_tag=$(git_tag)                                  \
	        commit_hash=$(commit_hash)                          \
	        commit_timestamp=$(commit_timestamp)                \
	        ./hack/build.sh ./...                               \
	    "

run: # @HELP builds and runs locally using native Go (no Docker)
	go run . serve

# ── Native Go targets (no Docker required) ────────────────────────────────────

.PHONY: test
test: # @HELP runs tests with race detector and coverage (native Go)
	go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | grep -E "^total:" | awk '{print "Coverage: " $$3}'

.PHONY: test-short
test-short: # @HELP runs tests without long-running cases (native Go)
	go test -short ./...

.PHONY: coverage-html
coverage-html: test # @HELP opens coverage report in browser
	go tool cover -html=coverage.out

.PHONY: vet
vet: # @HELP runs go vet
	go vet ./...

.PHONY: lint
lint: # @HELP runs golangci-lint (install: https://golangci-lint.run)
	golangci-lint run ./... --timeout=5m

.PHONY: vulncheck
vulncheck: # @HELP runs govulncheck for known vulnerabilities
	@which govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

.PHONY: check
check: vet test # @HELP runs vet + tests (native Go quality check)

.PHONY: tidy
tidy: # @HELP tidies and verifies Go modules
	go mod tidy
	go mod verify

# ── Docker build (native go binary inside image) ──────────────────────────────

.PHONY: docker-build
docker-build: # @HELP builds a single-arch Docker image using multi-stage Dockerfile
	DOCKER_BUILDKIT=1 docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg BUILD_DATE=$(commit_timestamp) \
	  --build-arg GIT_COMMIT=$(commit_hash) \
	  -t $(DOCKER_IMAGE):$(TAG) \
	  -f Dockerfile .

.PHONY: docker-build-push
docker-build-push: # @HELP builds and pushes multi-arch image via buildx

fmt: # @HELP Formats project source codes
fmt: $(BUILD_DIRS)
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    $(BUILD_IMAGE)                                          \
	    ./hack/fmt.sh $(EXCLUDE_DIRS)

verify-fmt: fmt
	@if !(git diff --exit-code HEAD); then 						\
		echo "formatting to source code is out of date"; exit 1;	\
	fi

# Example: make shell CMD="-c 'date > datefile'"
shell: # @HELP launches a shell in the containerized build environment
shell: $(BUILD_DIRS)
	@echo "launching a shell in the containerized build environment $(BUILD_IMAGE)"
	@docker run                                                 \
	    -ti                                                     \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    $(BUILD_IMAGE)                                          \
	    /bin/sh $(CMD)

start-server:
	go run ./cmd/grpc/server/main.go

start-client:
	go run ./cmd/grpc/client/main.go


mockgen: # @HELP Generate mock implementations for repo & database interfaces
mockgen:
	@echo "Generating mock implementations for repo & database interfaces"
	@docker run                                                 \
		-i                                                      \
		--rm                                                    \
		-u $$(id -u):$$(id -g)                                  \
		-v $$(pwd):/src                                         \
		-w /src                                                 \
		-v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
		-v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
		-v $$(pwd)/.go/cache:/.cache                            \
		--env HTTP_PROXY=$(HTTP_PROXY)                          \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                        \
		$(BUILD_IMAGE)                                          \
		/bin/bash -c " \
			$(foreach repo,$(REPO_PKGs), \
				mockgen -source=repos/$(repo).go -destination=repos/$(repo)/$(repo)_mock.go -package=$(repo); \
			) \
			$(foreach db,$(DB_TYPEs), \
				mockgen -source=infra/database/$(db)/database.go -destination=infra/database/$(db)/mock/mock.go -package=mock; \
			) \
		"

verify-mockgen: mockgen
	@if !(git diff --exit-code HEAD); then 						\
		echo "mock implementations for repo & database interfaces are out of date";	exit 1; 		\
	fi

modules: # @HELP Update module dependencies
modules: $(BUILD_DIRS)
	@echo "Updating go dependencies"
	@sudo chown -R $$(id -u):$$(id -g) $$(pwd)/.go
	@docker run                                                 \
		-i                                                      \
		--rm                                                    \
		-u $$(id -u):$$(id -g)                                  \
		-v $$(pwd):/src                                         \
		-w /src                                                 \
		-v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
		-v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
		-v $$(pwd)/.go/cache:/.cache                            \
		--env HTTP_PROXY=$(HTTP_PROXY)                          \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                        \
		$(BUILD_IMAGE)                                          \
		/bin/bash -c "											\
			go mod tidy && go mod vendor						\
		"

verify-modules: modules
	@if !(git diff --exit-code HEAD); then 						\
		echo "go module files are out of date";	exit 1; 		\
	fi


proto-gen: # @HELP Generate database protobuf codes
proto-gen:
	@echo "Generating database protobuf codes"
	@docker run                                                 \
		-i                                                      \
		--rm                                                    \
		-u $$(id -u):$$(id -g)                                  \
		-v $$(pwd):/src                                         \
		-w /src                                                 \
		-v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
		-v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
		-v $$(pwd)/.go/cache:/.cache                            \
		--env HTTP_PROXY=$(HTTP_PROXY)                          \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                        \
		$(BUILD_IMAGE)                                          \
		/bin/bash -c "	\
			protoc -I=/usr/include \
			--go_out=. --go_opt=module=github.com/masudur-rahman/pawsitively-purrfect \
			--go-grpc_out=. --go-grpc_opt=module=github.com/masudur-rahman/pawsitively-purrfect \
			-I=. proto/database/*.proto \
		"

verify-proto-gen: proto-gen
	@if !(git diff --exit-code HEAD); then 						\
		echo "database protobuf codes are out of date";	exit 1; 		\
	fi

schema-gen: # @HELP Generate GraphQL schema from graphql server
schema-gen:
	@get-graphql-schema http://pawsitively.purrfect:62783/graphql > hack/graphql/schema-generated.graphql

doc-gen: # @HELP Generate GraphQL docs based on generated schema
doc-gen:
	@npx spectaql hack/graphql/spectaql.yaml
	@sed -i '' 's#"images/favicon.png"#"https://lh3.googleusercontent.com/drive-viewer/AFGJ81oMoDuXwVfg4bTCqg0Q71sBPHDpHiTxOk6T2hlIJuUfHRCpdA-xeTmWQ6H58-wa8l6imLvASyQJfEEU0l3vgjFiLCwnNQ=s2560"#g' templates/index.html
	@sed -i '' 's#"images/logo.png"#"https://lh5.googleusercontent.com/9ZQCJ7yj0nccSqTTk-euc5Q7qzc5uKrsoNBD0zJ6trV-GSs7t68f-ZlxqEeKyglihTA=w2400"#g' templates/index.html
	@mv templates/index.html templates/docs.tmpl
	@graphql-markdown http://pawsitively.purrfect:62783/graphql > graphql.md

gen: proto-gen mockgen

verify: verify-modules verify-fmt

CONTAINER_DOTFILES = $(foreach bin,$(BIN),.container-$(subst /,_,$(REGISTRY)/$(bin))-$(TAG))

container containers: # @HELP builds containers for one platform ($OS/$ARCH)
container containers: $(CONTAINER_DOTFILES)
	@echo "container: $(DOCKER_IMAGE):$(TAG)"

# Each container-dotfile target can reference a $(BIN) variable.
# This is done in 2 steps to enable target-specific variables.
$(foreach bin,$(BIN),$(eval $(strip                                 \
    .container-$(subst /,_,$(REGISTRY)/$(bin))-$(TAG): BIN = $(bin)  \
)))
$(foreach bin,$(BIN),$(eval                                         \
    .container-$(subst /,_,$(REGISTRY)/$(bin))-$(TAG): bin/$(OS)_$(ARCH)/$(bin)$(BIN_EXTENSION) Dockerfile.in  \
))
# This is the target definition for all container-dotfiles.
# These are used to track build state in hidden files.
$(CONTAINER_DOTFILES):
	@sed                                          \
	    -e 's|{ARG_BIN}|$(BIN)$(BIN_EXTENSION)|g' \
	    -e 's|{ARG_ARCH}|$(ARCH)|g'               \
	    -e 's|{ARG_OS}|$(OS)|g'                   \
	    -e 's|{ARG_FROM}|$(BASEIMAGE)|g'          \
	    Dockerfile.in > .dockerfile-$(BIN)-$(OS)_$(ARCH)
	@docker buildx build --platform $(OS)/$(ARCH) --load --pull -t $(REGISTRY)/$(BIN):$(TAG) -f .dockerfile-$(BIN)-$(OS)_$(ARCH) .
	@docker images -q $(REGISTRY)/$(BIN):$(TAG) > $@
	@echo

push: # @HELP pushes the container for one platform ($OS/$ARCH) to the defined registry
push: $(CONTAINER_DOTFILES)
	docker push $(DOCKER_IMAGE):$(TAG);  \

manifest-list: # @HELP builds a manifest list of containers for all platforms
manifest-list: all-push
	platforms=$$(echo $(ALL_PLATFORMS) | sed 's/ /,/g');  \
	manifest-tool                                         \
		push from-args                                    \
		--platforms "$$platforms"                         \
		--template $(DOCKER_IMAGE):$(VERSION)_OS_ARCH  \
		--target $(DOCKER_IMAGE):$(VERSION)

.PHONY: docker-manifest
docker-manifest:
	docker manifest rm $(DOCKER_IMAGE):$(VERSION) | true
	docker manifest create -a $(DOCKER_IMAGE):$(VERSION) $(foreach PLATFORM,$(ALL_PLATFORMS),$(DOCKER_IMAGE):$(VERSION)_$(subst /,_,$(PLATFORM)))
	$(foreach PLATFORM,$(ALL_PLATFORMS), \
		docker manifest annotate $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):$(VERSION)_$(subst /,_,$(PLATFORM)) --arch $(lastword $(subst /, ,$(PLATFORM))); \
	)
	docker manifest push $(DOCKER_IMAGE):$(VERSION)

#docker-manifest: docker-manifest-PROD docker-manifest-DBG
#docker-manifest-%:
#	docker manifest create -a $(DOCKER_IMAGE):$(VERSION)_$* $(foreach PLATFORM,$(ALL_PLATFORMS),$(DOCKER_IMAGE):$(VERSION)_$*_$(subst /,_,$(PLATFORM)))
#	docker manifest push $(IMAGE):$(VERSION_$*)


.PHONY: release
release:
	@$(MAKE) all-push docker-manifest --no-print-directory

docker-build-push: # @HELP builds and pushes multi-arch image via buildx (legacy entry)
	docker buildx build --platform linux/amd64,linux/arm64 --output "type=image,push=true" --tag $(DOCKER_IMAGE):$(VERSION) --builder builder .

version: # @HELP outputs the version string
version:
	@echo "Application Version Information"
	@echo "==============================="
	@echo ""
	@echo "Application Name:    $(APP_NAME)"
	@echo ""
	@echo "Version Details:"
	@echo "    Version:            $(VERSION)"
	@echo "    Version Strategy:   $(version_strategy)"
	@echo ""
	@echo "Git Information:"
	@echo "    Git Tag:            $(git_tag)"
	@echo "    Git Branch:         $(git_branch)"
	@echo "    Commit Hash:        $(commit_hash)"
	@echo "    Commit Timestamp:   $(commit_timestamp)"
	@echo ""
	@echo "Build Environment:"
	@echo "    Go Version:         $(shell go version | cut -d " " -f 3)"
	@echo "    Compiler:           $(shell go env CC)"
	@echo "    Platform:           $(OS)/$(ARCH)"

test: # @HELP runs tests, as defined in ./hack/test.sh
test: $(BUILD_DIRS)
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    $(BUILD_IMAGE)                                          \
	    /bin/bash -c "                                          \
	        ARCH=$(ARCH)                                        \
	        OS=$(OS)                                            \
	        VERSION=$(VERSION)                                  \
	        ./hack/test.sh $(SRC_DIRS)                         \
	    "

$(BUILD_DIRS):
	@mkdir -p $@

clean: # @HELP removes built binaries and temporary files
clean: container-clean bin-clean

container-clean:
	rm -rf .container-* .dockerfile-*

bin-clean:
	rm -rf .go bin

help: # @HELP prints this message
help:
	@echo "VARIABLES:"
	@echo "  BIN = $(BIN)"
	@echo "  OS = $(OS)"
	@echo "  ARCH = $(ARCH)"
	@echo "  REGISTRY = $(REGISTRY)"
	@echo
	@echo "TARGETS:"
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST)    \
	    | awk '                                   \
	        BEGIN {FS = ": *# *@HELP"};           \
	        { printf "  %-30s %s\n", $$1, $$2 };  \
	    '
