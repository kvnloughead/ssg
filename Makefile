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
init: deps lint/install hooks

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

## dev: build binary, generate site, and serve (no watch)
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
# TESTING AND FORMATING
# ============================================================

## test: run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./...

## test/cover: run tests with coverage
.PHONY: test/cover
test/cover:
	@echo "Running tests with coverage..."
	@go test -cover ./...

## format: format code
.PHONY: format
format:
	@echo "Formatting Go code..."
	@go fmt ./...

## format/check: check if code is properly formatted
.PHONY: format/check
format/check:
	@echo "Checking Go formatting..."
	@files=$$(find . -name "*.go" -not -path "./vendor/*"); \
	if [ -n "$$files" ]; then \
		unformatted=$$(gofmt -l $$files); \
		if [ -n "$$unformatted" ]; then \
			echo "Code is not properly formatted. Run 'make format' to fix."; \
			echo "Unformatted files: $$unformatted"; \
			exit 1; \
		fi; \
	fi

# ============================================================
# LINTING AND SECURITY
# ============================================================

## lint/install: install linting tools
.PHONY: lint/install
lint/install:
	@echo "Installing linting tools (staticcheck and gosec)..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@if ! command -v djlint > /dev/null; then \
		echo "Warning: djlint not found. Install with: brew install djlint"; \
	fi
	@if ! command -v vnu > /dev/null; then \
		echo "Warning: vnu HTML validator not found. Install with: brew install vnu"; \
	fi

## lint: run static analysis
.PHONY: lint
lint: lint/templates validate/html
	@echo "Running staticcheck..."
	@staticcheck ./...

## security: run security analysis. This is a CLI that interacts with the user's local files, so G304 is excluded.
.PHONY: security
security:
	@echo "Running gosec security analysis..."
	@echo "This is a CLI the interacts with the users local files, so G304 is excluded."
	@gosec -exclude=G304 ./...

## lint/templates: lint HTML templates with djlint
.PHONY: lint/templates
lint/templates:
	@echo "Linting templates..."
	@djlint --profile=golang templates/

## validate/html: validate HTML with vnu (ignores trailing slash warnings)
.PHONY: validate/html
validate/html: generate
	@echo "Validating HTML with vnu..."
	@vnu --skip-non-html public/ 2>&1 | grep -v "Trailing slash on void elements" || true

## format/templates/check: check templates formatting
.PHONY: format/templates/check
format/templates/check:
	@echo "Checking template formatting..."
	@djlint --profile=golang --check templates/

## format/templates: format HTML templates with djlint
.PHONY: format/templates
format/templates:
	@echo "Formatting templates..."
	@djlint --profile=golang --reformat templates/

# ============================================================
# CONTINUOUS INTEGRATION
# ============================================================

## ci/test: run the test job like CI
.PHONY: ci/test
ci/test: test/cover

## ci/lint: run the lint job like CI (static analysis + security)
.PHONY: ci/lint
ci/lint: lint security

## ci/format: run the format job like CI
.PHONY: ci/format
ci/format: format/check

## ci/local: run full CI pipeline locally
.PHONY: ci/local
ci/local: ci/test ci/lint ci/format
	@echo "✅ All CI checks passed!"

## ci/setup: install all CI dependencies
.PHONY: ci/setup
ci/setup: deps lint/install

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
