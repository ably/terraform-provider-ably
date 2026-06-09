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

# Acceptance tests against a REAL Control API. Requires ABLY_ACCOUNT_TOKEN (and
# optionally ABLY_URL, e.g. staging). Setting TF_ACC makes TestMain stand aside
# so the suite hits the real API instead of the fake.
testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m
