# Expense Tracker Bot — Makefile

# The binary to build.
APP_NAME = Expense Tracker
BIN := expense-tracker-bot

# Where to push the docker image.
REGISTRY ?= masudjuly02

# Git metadata for versioning.
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

DOCKER_IMAGE := $(REGISTRY)/$(BIN)
GO_VERSION   ?= 1.24

# ── Build ─────────────────────────────────────────────────────────────────────

all: # @HELP builds the binary
all: build

.PHONY: build
build: # @HELP compiles the binary
	go build -o bin/$(BIN) .

run: # @HELP builds and runs locally using native Go (no Docker)
	go run . serve

# ── Test & Quality ────────────────────────────────────────────────────────────

.PHONY: test
test: # @HELP runs tests with race detector and coverage
	go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | grep -E "^total:" | awk '{print "Coverage: " $$3}'

.PHONY: test-short
test-short: # @HELP runs tests without long-running cases
	go test -short ./...

.PHONY: coverage-html
coverage-html: test # @HELP opens coverage report in browser
	go tool cover -html=coverage.out

.PHONY: vet
vet: # @HELP runs go vet
	go vet ./...

.PHONY: lint
lint: # @HELP runs golangci-lint
	golangci-lint run ./... --timeout=5m

.PHONY: vulncheck
vulncheck: # @HELP runs govulncheck for known vulnerabilities
	@which govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

.PHONY: check
check: vet lint test # @HELP runs vet + lint + tests

.PHONY: tidy
tidy: # @HELP tidies and verifies Go modules
	go mod tidy
	go mod verify

.PHONY: verify
verify: tidy # @HELP verifies modules are tidy
	@if ! git diff --exit-code go.mod go.sum; then \
		echo "go.mod/go.sum are out of date; run 'go mod tidy'"; exit 1; \
	fi

# ── Docker ────────────────────────────────────────────────────────────────────

.PHONY: docker-build
docker-build: # @HELP builds a single-arch Docker image using multi-stage Dockerfile
	DOCKER_BUILDKIT=1 docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg BUILD_DATE=$(commit_timestamp) \
	  --build-arg GIT_COMMIT=$(commit_hash) \
	  -t $(DOCKER_IMAGE):$(VERSION) \
	  -f Dockerfile .

.PHONY: docker-build-push
docker-build-push: # @HELP builds and pushes multi-arch image via buildx
	docker buildx build --platform linux/amd64,linux/arm64 --output "type=image,push=true" --tag $(DOCKER_IMAGE):$(VERSION) --builder builder .

.PHONY: release
release: # @HELP builds and pushes multi-arch image for release
	@$(MAKE) docker-build-push --no-print-directory

# ── Misc ──────────────────────────────────────────────────────────────────────

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
	@echo "    Platform:           $(shell go env GOOS)/$(shell go env GOARCH)"

.PHONY: clean
clean: # @HELP removes built binaries and temporary files
	rm -rf bin coverage.out

help: # @HELP prints this message
help:
	@echo "VARIABLES:"
	@echo "  BIN = $(BIN)"
	@echo "  REGISTRY = $(REGISTRY)"
	@echo
	@echo "TARGETS:"
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST)    \
	    | awk '                                   \
	        BEGIN {FS = ": *# *@HELP"};           \
	        { printf "  %-30s %s\n", $$1, $$2 };  \
	    '
