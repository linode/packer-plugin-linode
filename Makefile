NAME=linode
BINARY=packer-plugin-${NAME}
COUNT?=1
UNIT_TEST_TARGET?=$(shell go list ./builder/...)
HASHICORP_PACKER_PLUGIN_SDK_VERSION?=$(shell go list -m github.com/hashicorp/packer-plugin-sdk | cut -d " " -f2)
PACKER_SDC_REPO ?= github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc
.DEFAULT_GOAL = dev

# install is an alias of dev
.PHONY: install
install: dev

.PHONY: dev
dev: build
	@mkdir -p ~/.packer.d/plugins/
	@mv ${BINARY} ~/.packer.d/plugins/${BINARY}

.PHONY: build
build: fmtcheck
	@go build -o ${BINARY}

.PHONY: test
test: dev fmtcheck acctest

.PHONY: install-packer-sdc
install-packer-sdc: ## Install packer sofware development command
	@go install ${PACKER_SDC_REPO}@${HASHICORP_PACKER_PLUGIN_SDK_VERSION}

.PHONY: plugin-check
plugin-check: install-packer-sdc build
	@packer-sdc plugin-check ${BINARY}

.PHONY: unit-test
unit-test: dev
	@go test -count $(COUNT) -v $(UNIT_TEST_TARGET) -timeout=10m

# int-test is an alias of acctest
.PHONY: int-test
int-test: acctest

.PHONY: acctest
acctest: dev
	@PACKER_ACC=1 go test -count $(COUNT) ./... -v -timeout=100m

.PHONY: generate
generate: install-packer-sdc
	@go generate ./...
	@rm -rf .docs
	@packer-sdc renderdocs -src "docs" -partials docs-partials/ -dst ".docs/"
	@./.web-docs/scripts/compile-to-webdocs.sh "." ".docs" ".web-docs" "linode"
	@rm -r ".docs"

.PHONY: fmtcheck
fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

.PHONY: lint
lint: fmtcheck
	@golangci-lint run

.PHONY: format
format:
	@gofumpt -w .

.PHONY: deps
deps: install-packer-sdc
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install mvdan.cc/gofumpt@latest

.PHONY: clean
clean:
	@rm -rf .docs
	@rm -rf ./packer-plugin-linode
