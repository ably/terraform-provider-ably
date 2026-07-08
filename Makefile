TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=github.com
NAMESPACE=ably
NAME=ably
BINARY=terraform-provider-${NAME}
VERSION=1.0.0
OS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH_NATIVE := $(shell uname -m)
ARCH_MAPPED := $(shell echo "$(ARCH_NATIVE)" | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/' -e 's/armv7l/arm/')
ifneq ($(ARCH_MAPPED),$(filter amd64 arm64 arm,$(ARCH_MAPPED)))
$(error Unsupported architecture: $(ARCH_NATIVE))
endif
ARCH ?= $(ARCH_MAPPED)
OS_ARCH=${OS}_${ARCH}

default: install

build:
	go build -o ${BINARY}

release:
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_darwin_arm64
	GOOS=freebsd GOARCH=386   go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm   go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux   GOARCH=386   go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux   GOARCH=amd64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux   GOARCH=arm64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_linux_arm64
	GOOS=linux   GOARCH=arm   go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386   go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386   go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

# Hermetic test loop: unit tests plus the full acceptance suite run against an
# in-process fake Control API (see internal/provider/fake_control_api_test.go).
# No Ably credentials or network access required, safe to run on every change
# and in CI on forks. This is the loop an AI agent should run.
test:
	go test $(TEST) $(TESTARGS) -timeout=15m

# Acceptance tests against a REAL Control API. Requires ABLY_ACCOUNT_TOKEN and
# an explicit ABLY_URL (e.g. staging): TestMain refuses a real run that would
# fall back to the production URL implicitly. Setting TF_ACC makes TestMain
# stand aside so the suite hits the real API instead of the fake.
testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

# Regenerate Terraform schema/model code from the vendored Control API spec
# (codegen/control-api.yaml) using HashiCorp's tech-preview codegen tools. The
# generated code lands in internal/provider/codegen/. See codegen/README.md for
# how to refresh the spec and the current scope/caveats.
generate:
	# Track A: simple resources (app, namespace, queue) from the OpenAPI spec.
	go run github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi@v0.3.0 generate --config codegen/generator_config.yml --output codegen/spec.json codegen/control-api.yaml
	go run github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework@v0.4.1 generate resources --input codegen/spec.json --output internal/provider/codegen
	# Track B: rule families from the in-repo control types (the OpenAPI oneOf
	# union can't be generated, so we reflect the control rule structs instead).
	go run ./codegen/ruletypesgen
	go run github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework@v0.4.1 generate resources --input codegen/rules_spec.json --output internal/provider/codegen
	gofmt -w internal/provider/codegen

# Refresh the vendored Control API spec from the public ably/docs repo and
# re-apply any local fixes (codegen/spec-fixes.patch, if present), then
# regenerate. Never copy the upstream spec over codegen/control-api.yaml by
# hand: that silently reverts the fixes and the generators skip the affected
# attributes without erroring. Pass SPEC_SRC=<path> to use a local ably/docs
# checkout instead of fetching from GitHub.
SPEC_URL=https://raw.githubusercontent.com/ably/docs/main/static/open-specs/control-v1.yaml
refresh-spec:
ifdef SPEC_SRC
	cp $(SPEC_SRC) codegen/control-api.yaml
else
	curl -fsSL $(SPEC_URL) -o codegen/control-api.yaml
endif
	if [ -f codegen/spec-fixes.patch ]; then git apply codegen/spec-fixes.patch; fi
	$(MAKE) generate

.PHONY: build release install test testacc generate refresh-spec
