# Khorcha-Pati — Makefile

# The binary to build.
APP_NAME = Khorcha-Pati
BIN := khorcha-pati

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

# Prebuilt runtime base images carry the heavy fonts/wkhtmltopdf/chromium layers.
# Override BASE_REF_* with an immutable digest (…-base@sha256:…) to pin in CI.
BASE_IMAGE        := $(REGISTRY)/$(BIN)-base
BASE_REF_WK       ?= $(BASE_IMAGE):wkhtmltopdf
BASE_REF_CHROMEDP ?= $(BASE_IMAGE):chromedp

# Local dev: allowed host + direct API base for the web dev server (override per run if needed).
DEV_WEB_HOST ?= khorchapati.mrahman.xyz
DEV_API      ?= https://khorchapati-api.mrahman.xyz

# ── Build ─────────────────────────────────────────────────────────────────────

all: # @HELP builds the binary
all: build

.PHONY: build
build: # @HELP compiles the binary
	go build -o bin/$(BIN) .

run: # @HELP builds and runs locally using native Go (no Docker)
	go run . serve

.PHONY: dev
dev: # @HELP runs backend + web dev server together with prefixed logs (ctrl+c stops both)
	@trap 'kill 0' INT TERM; \
	( go run . serve 2>&1 | sed -u 's/^/[api] /' ) & \
	( cd web && VITE_ALLOWED_HOSTS=$(DEV_WEB_HOST) VITE_API_BASE=$(DEV_API) npm run dev 2>&1 | sed -u 's/^/[web] /' ) & \
	wait

# ── Format ───────────────────────────────────────────────────────────────────

.PHONY: fmt
fmt: # @HELP formats Go and shell source files
	@which goimports-reviser >/dev/null 2>&1 || go install github.com/incu6us/goimports-reviser/v3@latest
	goimports-reviser -recursive -company-prefixes=github.com/masudur-rahman -imports-order=std,project,company,general,blanked -format -excludes vendor ./...
	@which shfmt >/dev/null 2>&1 || go install mvdan.cc/sh/v3/cmd/shfmt@latest
	find . -path ./vendor -prune -o -name '*.sh' -exec shfmt -l -w -ci -i 4 {} \;

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
# PDF_GENERATOR selects the Dockerfile target: wkhtmltopdf (default) or chromedp
PDF_GENERATOR ?= wkhtmltopdf
DOCKER_TAG_SUFFIX := $(if $(filter chromedp,$(PDF_GENERATOR)),-chromedp,)

.PHONY: docker-build
docker-build: # @HELP builds a single-arch Docker image (PDF_GENERATOR=wkhtmltopdf|chromedp)
	DOCKER_BUILDKIT=1 docker build \
	  --target $(PDF_GENERATOR) \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg BUILD_DATE=$(commit_timestamp) \
	  --build-arg GIT_COMMIT=$(commit_hash) \
	  --build-arg BASE_WK=$(BASE_REF_WK) \
	  --build-arg BASE_CHROMEDP=$(BASE_REF_CHROMEDP) \
	  $(DOCKER_CACHE_ARGS) \
	  -t $(DOCKER_IMAGE):$(VERSION)$(DOCKER_TAG_SUFFIX) \
	  -f Dockerfile .

.PHONY: docker-run
docker-run: # @HELP run container built from latest changes
	@if [ -z "$$(docker images -q $(DOCKER_IMAGE):$(VERSION)$(DOCKER_TAG_SUFFIX))" ]; then \
		echo "Image not found. Building..."; \
		$(MAKE) docker-build; \
	else \
		echo "Image $(DOCKER_IMAGE):$(VERSION)$(DOCKER_TAG_SUFFIX) already exists. Skipping build."; \
	fi
	docker run \
	  --rm \
	  --env-file .env \
	  --volume $(CURDIR)/.configs/.khorcha-pati-docker.yaml:/app/.configs/.khorcha-pati.yaml \
	  $(DOCKER_IMAGE):$(VERSION)$(DOCKER_TAG_SUFFIX)

.PHONY: docker-build-web
docker-build-web: # @HELP builds web frontend Docker image for current platform
	docker build -t $(DOCKER_IMAGE)-web:$(VERSION) web/

.PHONY: docker-push-web
docker-push-web: # @HELP pushes web frontend Docker image to registry
	docker buildx build --platform linux/amd64,linux/arm64 \
	  --output "type=image,push=true" \
	  --tag $(DOCKER_IMAGE)-web:$(VERSION) web/

.PHONY: docker-compose-up
docker-compose-up: # @HELP runs backend + frontend via Docker Compose
	docker compose up --build

.PHONY: docker-build-push
docker-build-push: # @HELP builds and pushes multi-arch image (PDF_GENERATOR=wkhtmltopdf|chromedp)
	docker buildx build --platform linux/amd64,linux/arm64 \
	  --target $(PDF_GENERATOR) \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg BUILD_DATE=$(commit_timestamp) \
	  --build-arg GIT_COMMIT=$(commit_hash) \
	  --build-arg BASE_WK=$(BASE_REF_WK) \
	  --build-arg BASE_CHROMEDP=$(BASE_REF_CHROMEDP) \
	  $(DOCKER_CACHE_ARGS) \
	  --output "type=image,push=true" \
	  --tag $(DOCKER_IMAGE):$(VERSION)$(DOCKER_TAG_SUFFIX) \
	  -f Dockerfile .

.PHONY: base-build-local
base-build-local: # @HELP builds the runtime base locally (single-arch, not pushed) for CI/local image verification
	DOCKER_BUILDKIT=1 docker build \
	  --target $(PDF_GENERATOR) \
	  $(DOCKER_CACHE_ARGS) \
	  -t $(if $(filter chromedp,$(PDF_GENERATOR)),$(BASE_REF_CHROMEDP),$(BASE_REF_WK)) \
	  -f Dockerfile.base .

.PHONY: ensure-base
ensure-base: # @HELP builds the base images only if they are missing from the registry (self-heals a cold release)
	@if docker manifest inspect $(BASE_REF_WK) >/dev/null 2>&1 && \
	    docker manifest inspect $(BASE_REF_CHROMEDP) >/dev/null 2>&1; then \
		echo "Base images present — skipping base build."; \
	else \
		echo "Base image(s) missing — building them first..."; \
		$(MAKE) base-build-push --no-print-directory; \
	fi

.PHONY: base-build-push
base-build-push: # @HELP builds and pushes the prebuilt runtime base images (wkhtmltopdf + chromedp)
	docker buildx build --platform linux/amd64,linux/arm64 \
	  --target wkhtmltopdf \
	  $(DOCKER_CACHE_ARGS) \
	  --output "type=image,push=true" \
	  --tag $(BASE_REF_WK) \
	  -f Dockerfile.base .
	docker buildx build --platform linux/amd64,linux/arm64 \
	  --target chromedp \
	  $(DOCKER_CACHE_ARGS) \
	  --output "type=image,push=true" \
	  --tag $(BASE_REF_CHROMEDP) \
	  -f Dockerfile.base .

.PHONY: release
release: # @HELP builds and pushes both wkhtmltopdf and chromedp backend images
	@$(MAKE) docker-build-push PDF_GENERATOR=wkhtmltopdf --no-print-directory
	@$(MAKE) docker-build-push PDF_GENERATOR=chromedp --no-print-directory

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
