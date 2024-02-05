NAME=linode
BINARY=packer-plugin-${NAME}
GOFMT_FILES?=$$(find . -name '*.go')
COUNT?=1
TEST?=$(shell go list ./builder/...)
HASHICORP_PACKER_PLUGIN_SDK_VERSION?=$(shell go list -m github.com/hashicorp/packer-plugin-sdk | cut -d " " -f2)

.PHONY: dev

build: fmtcheck
	@go build -o ${BINARY}

dev: build
	@mkdir -p ~/.packer.d/plugins/
	@mv ${BINARY} ~/.packer.d/plugins/${BINARY}

test: dev fmtcheck
	@PACKER_ACC=1 go test -count $(COUNT) ./... -v -timeout=100m

install-packer-sdc: ## Install packer sofware development command
	@go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@${HASHICORP_PACKER_PLUGIN_SDK_VERSION}

plugin-check: install-packer-sdc build
	@packer-sdc plugin-check ${BINARY}

unit-test: dev
	@PACKER_ACC=1 go test -count $(COUNT) -v $(TEST) -timeout=10m

int-test: dev
	@go test -v test/integration/e2e_test.go

generate: install-packer-sdc
	@go generate ./...
	@rm -rf .docs
	@packer-sdc renderdocs -src "docs" -partials docs-partials/ -dst ".docs/"
	@./.web-docs/scripts/compile-to-webdocs.sh "." ".docs" ".web-docs" "linode"
	@rm -r ".docs"

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

lint: fmtcheck
	golangci-lint run --timeout 15m0s

fmt:
	gofmt -w $(GOFMT_FILES)
	gofumpt -w .
