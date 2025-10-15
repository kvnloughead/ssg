-include .env

# ============================================================
# BUILD AND GENERATE
# ============================================================

## build: build the ssg binary
.PHONY: build
build:
	@echo "Building SSG..."
	@go build -o bin/ssg ./cmd/ssg

## deps: install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go get gopkg.in/yaml.v3
	@go get github.com/yuin/goldmark

## init: initialize project (run once)
.PHONY: init
init: ci/setup hooks

# ============================================================
# SITE GENERATION
# ============================================================

## generate: generate the static site to public/
.PHONY: generate
generate: build
	@echo "Generating static site..."
	@./bin/ssg build

## serve: serve the generated site locally
.PHONY: serve
serve:
	@echo "Serving site..."
	@./bin/ssg serve

## new: create a new post (requires TITLE variable)
.PHONY: new
new:
ifndef TITLE
	@echo "Error: TITLE variable is required"
	@echo "Usage: make new TITLE=\"Your Post Title\""
	@exit 1
endif
	@./bin/ssg new --title "$(TITLE)"

## run/dev: build binary, generate site, and serve (no watch)
.PHONY: run/dev
run/dev: build generate serve

## run/air: run with air for live reload (watches all files)
.PHONY: run/air
run/air:
	@if ! command -v air > /dev/null; then \
		echo "Installing air for live reload..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@air

# ============================================================
# CONTINUOUS INTEGRATION
# ============================================================

## ci/setup: install all CI dependencies
.PHONY: ci/setup
ci/setup: deps
	@echo "Installing linting tools (staticcheck and gosec)..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Installing platform-specific tools..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		if ! command -v act > /dev/null; then \
			echo "Installing act..."; \
			brew install act; \
		fi; \
		if ! command -v djlint > /dev/null; then \
			echo "Installing djlint..."; \
			brew install djlint; \
		fi; \
		if ! command -v vnu > /dev/null; then \
			echo "Installing vnu..."; \
			brew install vnu; \
		fi; \
	elif [ "$$(uname)" = "Linux" ]; then \
		if ! command -v djlint > /dev/null; then \
			echo "Installing djlint via pip..."; \
			pip install djlint || pip3 install djlint; \
		fi; \
		if ! command -v vnu > /dev/null; then \
			echo "Installing vnu via npm..."; \
			npm install -g vnu-jar; \
		fi; \
	else \
		echo "Warning: Unsupported platform. Please install djlint and vnu manually."; \
	fi

## ci/test: run tests with coverage like CI
.PHONY: ci/test
ci/test:
	@echo "Running tests with coverage..."
	@go test -cover ./...

## ci/lint: run linting like CI (static analysis + security + templates + HTML validation)
.PHONY: ci/lint
ci/lint:
	@echo "Linting templates..."
	@djlint --profile=golang templates/
	@echo "Running staticcheck..."
	@staticcheck ./...
	@echo "Running gosec security analysis..."
	@echo "This is a CLI that interacts with the user's local files, so G304 is excluded."
	@gosec -exclude=G304 ./...

## ci/validate: validate generated HTML with vnu
.PHONY: ci/validate
ci/validate: generate
	@echo "Validating HTML with vnu..."
	@vnu --skip-non-html public/ 2>&1 | grep -v "Trailing slash on void elements" || true

## ci/format: check formatting like CI
.PHONY: ci/format
ci/format:
	@echo "Checking Go formatting..."
	@files=$$(find . -name "*.go" -not -path "./vendor/*"); \
	if [ -n "$$files" ]; then \
		unformatted=$$(gofmt -l $$files); \
		if [ -n "$$unformatted" ]; then \
			echo "Code is not properly formatted. Run 'make fix' to fix."; \
			echo "Unformatted files: $$unformatted"; \
			exit 1; \
		fi; \
	fi
	@echo "Checking template formatting..."
	@djlint --profile=golang --indent 2 --check templates/

## ci/local: run full CI pipeline locally
.PHONY: ci/local
ci/local: ci/test ci/lint ci/validate ci/format
	@echo "✅ All CI checks passed!"

# ============================================================
# FORMATTING AND FIXING
# ============================================================

## fix: auto-fix formatting issues
.PHONY: format/fix
fix:
	@echo "Formatting Go code and templates..."
	@go fmt ./...
	@echo "Formatting templates..."
	@djlint --profile=golang --reformat templates/

# ============================================================
# GIT HOOKS
# ============================================================

## hooks: enable local git hooks to prevent pushing if tests fail
hooks:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-push

# ============================================================
# HELPERS
# ============================================================

## help: print this help message
.PHONY: help
help:
	@echo "\nUsage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]
